package sfsturbo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	log "k8s.io/klog/v2"
)

// KubeClient wraps the Kubernetes clientset and provides helper methods to
// read PVC annotations and PV volumeAttributes. The driver uses this to access
// PVC-level metadata (annotations) that are not forwarded by the standard
// csi-provisioner.
type KubeClient struct {
	clientset kubernetes.Interface
}

// NewKubeClient creates a new KubeClient. It first tries the in-cluster config;
// if that fails (e.g. running outside a cluster), it falls back to the
// kubeconfig file pointed to by KUBECONFIG env var or ~/.kube/config.
func NewKubeClient() (*KubeClient, error) {
	cfg, err := rest.InClusterConfig()
	if err == nil {
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster clientset: %v", err)
		}
		log.Infof("Using in-cluster Kubernetes config")
		return &KubeClient{clientset: clientset}, nil
	}

	// Fall back to kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}
	cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset from kubeconfig: %v", err)
	}
	log.Infof("Using kubeconfig from %s", kubeconfig)
	return &KubeClient{clientset: clientset}, nil
}

// SubDirVolumeContext holds the parameters extracted from PVC annotations for
// the sub-directory provisioning mode.
type SubDirVolumeContext struct {
	ShareID       string
	Path          string
	ReclaimPolicy string
	QuotaEnabled  bool
}

// GetPVCAnnotations reads the PVC with the given name and namespace from the
// Kubernetes API and returns its annotations.
func (k *KubeClient) GetPVCAnnotations(pvcName, pvcNamespace string) (map[string]string, error) {
	if k == nil || k.clientset == nil {
		return nil, status.Error(codes.Internal, "Kubernetes client is not initialized")
	}
	pvc, err := k.clientset.CoreV1().PersistentVolumeClaims(pvcNamespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Failed to get PVC %s/%s: %v", pvcNamespace, pvcName, err)
	}
	return pvc.Annotations, nil
}

// GetSubDirVolumeContext reads the PVC annotations and extracts the
// sub-directory parameters. Returns nil (and no error) if the PVC does not
// have the volume-as annotation set to "absolute-path".
func (k *KubeClient) GetSubDirVolumeContext(pvcName, pvcNamespace string) (*SubDirVolumeContext, error) {
	annotations, err := k.GetPVCAnnotations(pvcName, pvcNamespace)
	if err != nil {
		return nil, err
	}
	return parseSubDirAnnotations(annotations), nil
}

// parseSubDirAnnotations extracts the sub-directory parameters from the given
// annotation map. Returns nil if the volume-as annotation is not set to
// "absolute-path". Validates that all required fields are present.
func parseSubDirAnnotations(annotations map[string]string) *SubDirVolumeContext {
	volumeAs, ok := annotations[VolumeAsAnnotation]
	if !ok || volumeAs != VolumeAsAbsolutePath {
		return nil
	}

	shareID := annotations[SfsTurboShareIDAnnotation]
	path := annotations[PathAnnotation]
	reclaimPolicy := annotations[ReclaimPolicyAnnotation]
	quotaEnabled := annotations[EnableDirQuotaAnnotation] == "true"

	return &SubDirVolumeContext{
		ShareID:       shareID,
		Path:          path,
		ReclaimPolicy: reclaimPolicy,
		QuotaEnabled:  quotaEnabled,
	}
}

// ValidateSubDirVolumeContext validates the sub-directory parameters and
// returns an error if any required field is missing or invalid.
func ValidateSubDirVolumeContext(ctx *SubDirVolumeContext) error {
	if ctx.ShareID == "" {
		return status.Errorf(codes.InvalidArgument,
			"annotation %s is required when %s is %s",
			SfsTurboShareIDAnnotation, VolumeAsAnnotation, VolumeAsAbsolutePath)
	}
	if ctx.Path == "" {
		return status.Errorf(codes.InvalidArgument,
			"annotation %s is required when %s is %s",
			PathAnnotation, VolumeAsAnnotation, VolumeAsAbsolutePath)
	}
	if len(ctx.Path) == 0 || ctx.Path[0] != '/' {
		return status.Errorf(codes.InvalidArgument,
			"annotation %s must be an absolute path (start with /), got: %s",
			PathAnnotation, ctx.Path)
	}
	if strings.Contains(ctx.Path, "#") {
		return status.Errorf(codes.InvalidArgument,
			"annotation %s must not contain '#' (used as the VolumeId separator), got: %s",
			PathAnnotation, ctx.Path)
	}
	if ctx.ReclaimPolicy == "" {
		// Default to retain-volume-only if not specified.
		ctx.ReclaimPolicy = ReclaimPolicyRetainVolumeOnly
	}
	if ctx.ReclaimPolicy != ReclaimPolicyRetainVolumeOnly && ctx.ReclaimPolicy != ReclaimPolicyDelete {
		return status.Errorf(codes.InvalidArgument,
			"annotation %s must be %q or %q, got: %s",
			ReclaimPolicyAnnotation, ReclaimPolicyRetainVolumeOnly, ReclaimPolicyDelete, ctx.ReclaimPolicy)
	}
	return nil
}

// GetPVVolumeContext reads the PV with the given name and returns the CSI
// volumeAttributes (volumeContext). This is used in DeleteVolume to retrieve
// the reclaim policy and other context stored during CreateVolume.
func (k *KubeClient) GetPVVolumeContext(pvName string) (map[string]string, error) {
	if k == nil || k.clientset == nil {
		return nil, status.Error(codes.Internal, "Kubernetes client is not initialized")
	}
	pv, err := k.clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pvName, metav1.GetOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Failed to get PV %s: %v", pvName, err)
	}
	if pv.Spec.CSI == nil {
		return nil, status.Errorf(codes.Internal,
			"PV %s has no CSI source", pvName)
	}
	return pv.Spec.CSI.VolumeAttributes, nil
}

// GetReclaimPolicyFromPV reads the PV with the given name and returns the
// reclaim policy from its CSI volumeAttributes. This is used in DeleteVolume
// to decide whether to delete the sub-directory.
func (k *KubeClient) GetReclaimPolicyFromPV(pvName string) (string, error) {
	attrs, err := k.GetPVVolumeContext(pvName)
	if err != nil {
		return "", err
	}
	reclaimPolicy, ok := attrs[VolumeContextReclaimPolicy]
	if !ok {
		// If the reclaim policy is not stored in the PV (e.g. old PV), default
		// to retain-volume-only for safety.
		log.Warningf("PV %s does not have %s in volumeAttributes, defaulting to %s",
			pvName, VolumeContextReclaimPolicy, ReclaimPolicyRetainVolumeOnly)
		return ReclaimPolicyRetainVolumeOnly, nil
	}
	return reclaimPolicy, nil
}

// GetReclaimPolicyByVolumeHandle lists all PVs and finds the one whose CSI
// volumeHandle matches the given volumeID. This is used in DeleteVolume when
// the PV name is not directly available (the standard csi-provisioner does not
// pass the PV name in DeleteVolume requests).
// Returns the reclaim policy from the PV's volumeAttributes, or
// ReclaimPolicyRetainVolumeOnly if the PV is not found or the reclaim policy
// is not stored.
func (k *KubeClient) GetReclaimPolicyByVolumeHandle(volumeID string) (string, error) {
	if k == nil || k.clientset == nil {
		return ReclaimPolicyRetainVolumeOnly, status.Error(codes.Internal, "Kubernetes client is not initialized")
	}
	pvs, err := k.clientset.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return ReclaimPolicyRetainVolumeOnly, status.Errorf(codes.Internal,
			"Failed to list PVs: %v", err)
	}
	for _, pv := range pvs.Items {
		if pv.Spec.CSI == nil {
			continue
		}
		if pv.Spec.CSI.VolumeHandle == volumeID {
			attrs := pv.Spec.CSI.VolumeAttributes
			reclaimPolicy, ok := attrs[VolumeContextReclaimPolicy]
			if !ok {
				log.Warningf("PV %s (volumeHandle=%s) does not have %s in volumeAttributes, defaulting to %s",
					pv.Name, volumeID, VolumeContextReclaimPolicy, ReclaimPolicyRetainVolumeOnly)
				return ReclaimPolicyRetainVolumeOnly, nil
			}
			return reclaimPolicy, nil
		}
	}
	// PV not found — it may have already been deleted by Kubernetes.
	// Default to retain-volume-only for safety.
	log.Infof("PV with volumeHandle %s not found, defaulting to %s", volumeID, ReclaimPolicyRetainVolumeOnly)
	return ReclaimPolicyRetainVolumeOnly, nil
}

// --- VolumeId encoding helpers ---
// The sub-directory mode encodes the share ID and path into the CSI VolumeId
// as "<shareID>#<path>". This keeps the VolumeId self-contained so that
// NodePublishVolume can decode it without needing a K8s API call.

// encodeVolumeID encodes a shareID and path into a single VolumeId string.
func encodeVolumeID(shareID, path string) string {
	return shareID + volumeIDSep + path
}

// decodeVolumeID splits a VolumeId into shareID and path. If the VolumeId does
// not contain the separator, it is a legacy VolumeId (plain share UUID) and
// (volumeID, "") is returned.
func decodeVolumeID(volumeID string) (shareID, path string) {
	for i := len(volumeID) - 1; i >= 0; i-- {
		if volumeID[i] == '#' {
			return volumeID[:i], volumeID[i+1:]
		}
	}
	return volumeID, ""
}

// isSubDirVolumeID returns true if the VolumeId contains the separator,
// indicating it uses the sub-directory mode.
func isSubDirVolumeID(volumeID string) bool {
	_, path := decodeVolumeID(volumeID)
	return path != ""
}

// EnsureKubeClient creates a KubeClient if the cloud config does not already
// provide one. This is called during driver initialization.
func EnsureKubeClient() *KubeClient {
	kc, err := NewKubeClient()
	if err != nil {
		log.Errorf("Failed to create Kubernetes client: %v. "+
			"Sub-directory mode will not be available.", err)
		return nil
	}
	log.Infof("Kubernetes client initialized successfully for sub-directory mode")
	return kc
}
