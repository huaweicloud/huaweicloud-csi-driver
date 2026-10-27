package framework

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// CreatePVC create PersistentVolumeClaim.
func CreatePVC(client kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) {
	ginkgo.By(fmt.Sprintf("Creating PersistentVolumeClaim(%s/%s)", pvc.Namespace, pvc.Name), func() {
		_, err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// UpdatePVC update PersistentVolumeClaim.
func UpdatePVC(client kubernetes.Interface, pvc *corev1.PersistentVolumeClaim) {
	ginkgo.By(fmt.Sprintf("Updateing PersistentVolumeClaim(%s/%s)", pvc.Namespace, pvc.Name), func() {
		_, err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Update(context.TODO(), pvc, metav1.UpdateOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
}

// GetPVC get PersistentVolumeClaim
func GetPVC(client kubernetes.Interface, namespace, name string) *corev1.PersistentVolumeClaim {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	return pvc
}

// RemovePVC delete PersistentVolumeClaim. Ignores not-found errors so it is safe
// to call when the PVC may have already been deleted by the test itself.
func RemovePVC(client kubernetes.Interface, namespace, name string) {
	ginkgo.By(fmt.Sprintf("Removing PersistentVolumeClaim(%s/%s)", namespace, name), func() {
		err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		}
	})
}

func WaitPVCPresentOnClusterFitWith(client kubernetes.Interface, namespace, name string,
	fit func(pvc *corev1.PersistentVolumeClaim) bool) {
	klog.Infof("Waiting for PVC(%s/%s) synced", namespace, name)
	gomega.Eventually(func() bool {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return fit(pvc)
	}, pollTimeout, pollInterval).Should(gomega.Equal(true))
}

// WaitPVCExpanded waits until the PVC capacity is updated to the requested size.
// If a resize failure condition is detected on the PVC (PersistentVolumeClaimResizing
// set to False), the failure reason and message are logged as warnings so they are
// visible in the test output. On timeout the last error message — which includes
// the resize failure details — is included in the test failure.
func WaitPVCExpanded(client kubernetes.Interface, namespace, name string, newSize int64) {
	klog.Infof("Waiting for PVC(%s/%s) expanded to %d bytes", namespace, name, newSize)
	var lastResizeErr string
	gomega.Eventually(func() string {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Sprintf("failed to get PVC(%s/%s): %v", namespace, name, err)
		}
		capacity := pvc.Status.Capacity.Storage().Value()
		if capacity == newSize {
			return ""
		}
		// Check for resize failure conditions and surface the actual error.
		for _, cond := range pvc.Status.Conditions {
			if cond.Type == corev1.PersistentVolumeClaimResizing && cond.Status == corev1.ConditionFalse {
				lastResizeErr = fmt.Sprintf("reason=%s, message=%s", cond.Reason, cond.Message)
				klog.Warningf("PVC(%s/%s) resize failed: %s", namespace, name, lastResizeErr)
			}
		}
		msg := fmt.Sprintf("PVC(%s/%s) capacity is %d, waiting for %d", namespace, name, capacity, newSize)
		if lastResizeErr != "" {
			msg += fmt.Sprintf(", last resize error: %s", lastResizeErr)
		}
		return msg
	}, pollTimeout, pollInterval).Should(gomega.BeEmpty())
}

// WaitPVCBound waits until the PVC is bound to a PV and returns the PV name.
//
// On timeout the failure message includes the PVC phase and the most recent
// ProvisioningFailed event (if any). This surfaces provisioning failures —
// e.g. an SFS Turbo quota error like "SFS.TURBO.0113 dir not create quota" —
// directly in the test output instead of a generic "expected not to be empty".
// This keeps provisioning regressions from being masked as a 600s Pod-timeout.
func WaitPVCBound(client kubernetes.Interface, namespace, name string) string {
	klog.Infof("Waiting for PVC(%s/%s) bound", namespace, name)
	gomega.Eventually(func() string {
		pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Sprintf("failed to get PVC(%s/%s): %v", namespace, name, err)
		}
		if pvc.Spec.VolumeName != "" {
			return ""
		}
		// Not bound yet — build a descriptive message that includes the PVC
		// phase and the latest ProvisioningFailed event so the root cause
		// (e.g. a quota provisioning error) is visible in the test failure.
		phase := string(pvc.Status.Phase)
		if phase == "" {
			phase = "Unknown"
		}
		msg := fmt.Sprintf("PVC(%s/%s) not bound yet, phase=%s", namespace, name, phase)
		if eventErr := latestProvisioningFailedEvent(client, namespace, name); eventErr != "" {
			msg += fmt.Sprintf(", provisioning failure: %s", eventErr)
		}
		return msg
	}, pollTimeout, pollInterval).Should(gomega.BeEmpty())

	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	return pvc.Spec.VolumeName
}

// latestProvisioningFailedEvent returns the message of the most recent
// ProvisioningFailed event for the given PVC, or "" if none is found. It is
// best-effort: errors fetching events are returned as part of the message.
func latestProvisioningFailedEvent(client kubernetes.Interface, namespace, pvcName string) string {
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=PersistentVolumeClaim,involvedObject.name=%s", pvcName),
	}
	events, err := client.CoreV1().Events(namespace).List(context.TODO(), opts)
	if err != nil {
		return fmt.Sprintf("failed to list events: %v", err)
	}
	var latest *corev1.Event
	for i := range events.Items {
		e := &events.Items[i]
		if e.Reason != "ProvisioningFailed" {
			continue
		}
		if latest == nil || e.LastTimestamp.After(latest.LastTimestamp.Time) {
			latest = e
		}
	}
	if latest == nil {
		return ""
	}
	return fmt.Sprintf("reason=%s, message=%s", latest.Reason, latest.Message)
}

// WaitPVDeleted waits until the PersistentVolume with the given name is removed
// from the cluster (reclaimPolicy=Delete path).
func WaitPVDeleted(client kubernetes.Interface, pvName string) {
	klog.Infof("Waiting for PV(%s) deleted", pvName)
	gomega.Eventually(func() bool {
		_, err := client.CoreV1().PersistentVolumes().Get(context.TODO(), pvName, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return true
		}
		return false
	}, pollTimeout, pollInterval).Should(gomega.Equal(true))
}
