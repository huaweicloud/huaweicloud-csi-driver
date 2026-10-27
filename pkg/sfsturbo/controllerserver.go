/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sfsturbo

import (
	"reflect"
	"strconv"

	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/fs"
	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/shares"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/sfsturbo/services"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
)

type controllerServer struct {
	Driver *SfsTurboDriver
}

const (
	defaultShareProto = "NFS"
	defaultShareType  = "STANDARD"
	minSizeInGiB      = 500
	maxSizeInGiB      = 32768
	expandStepInGiB   = 100

	// expandTypeHpc is the metadata.expand_type value used to create HPC file
	// systems (20MB/s/TiB ~ 1000MB/s/TiB). For these specs the SFS Turbo API
	// does not validate share_type and uses a different size range, so the
	// driver skips the legacy 500~32768 GiB checks and 100GiB expand step.
	expandTypeHpc      = "hpc"
	expandTypeHpcCache = "hpc_cache"
)

// resourcemode from SC
const (
	Availability                 = "availability"
	ShareType                    = "shareType"
	AutoCreateSecurityGroupRules = "autoCreateSecurityGroupRules"
	ExpandType                   = "expandType"
	HpcBw                        = "hpcBw"
)

func (cs *controllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {
	log.Infof("CreateVolume called with request %v", protosanitizer.StripSecrets(*req))
	cloud := cs.Driver.cloud

	name := req.GetName()
	capacityRange := req.GetCapacityRange()
	if err := createVolumeValidation(name, capacityRange); err != nil {
		return nil, err
	}

	parameters := req.GetParameters()

	// Check if the sub-directory mode is requested via PVC annotations.
	// The csi-provisioner with --extra-create-metadata injects the PVC name
	// and namespace into parameters; we use them to fetch the PVC and read
	// its annotations.
	subDirCtx, err := cs.tryGetSubDirContext(parameters)
	if err != nil {
		return nil, err
	}
	if subDirCtx != nil {
		return cs.createSubDirVolume(req, subDirCtx)
	}

	// --- Legacy mode: create a new SFS Turbo share ---
	sizeInGiB := int(utils.RoundUpSize(capacityRange.GetRequiredBytes(), common.GbByteSize))

	// Resolve the HPC parameters with loose inference: when hpcBw is set
	// without expandType, expandType is inferred as "hpc". When neither is
	// set, the request targets a legacy STANDARD/PERFORMANCE share.
	expandType, hpcBw := resolveExpandParams(parameters)
	isHpc := expandType == expandTypeHpc || expandType == expandTypeHpcCache

	// HPC shares leave capacity validation to the SFS Turbo API (different
	// size range and 1.2TiB multiple). Legacy shares keep the original
	// 500~32768 GiB range and auto-raising below the minimum.
	if !isHpc {
		var err error
		sizeInGiB, err = validateLegacyCreateSize(sizeInGiB)
		if err != nil {
			return nil, err
		}
	}

	var accessibleTopology []*csi.Topology
	// First check if volAvailability is already specified, if not get preferred from Topology
	// Required, incase vol AZ is different from node AZ
	volumeAz := parameters[Availability]

	accessibleTopology, err = cs.getAccessibleTopology(req, volumeAz)
	if err != nil {
		return nil, err
	}

	shareType := parameters[ShareType]
	if len(shareType) == 0 {
		shareType = defaultShareType
	}

	autoCreateSGRules, err := getAutoCreateSGRules(parameters)
	if err != nil {
		return nil, err
	}

	metadata := shares.Metadata{
		AutoCreateSecurityGroupRules: autoCreateSGRules,
	}
	// Set HPC metadata only when an HPC share is requested.
	if isHpc {
		metadata.ExpandType = expandTypeHpc
		metadata.HpcBw = hpcBw
	}

	createShareOpts := shares.Share{
		Name:             name,
		ShareProto:       defaultShareProto,
		ShareType:        shareType,
		Size:             sizeInGiB,
		AvailabilityZone: volumeAz,
		VpcID:            cs.Driver.cloud.Vpc.ID,
		SubnetID:         cs.Driver.cloud.Vpc.SubnetID,
		SecurityGroupID:  cs.Driver.cloud.Vpc.SecurityGroupID,
		Metadata:         metadata,
	}
	log.Infof("CreateVolume creating param: %v", protosanitizer.StripSecrets(createShareOpts))
	// Check if there are any volumes with the same name
	share, err := checkVolumeExists(cloud, createShareOpts, isHpc)
	if err != nil {
		return nil, err
	}

	if share != nil {
		log.Infof("CreateVolume already exist share: %v", protosanitizer.StripSecrets(share))
		size, err := strconv.ParseFloat(share.Size, 64)
		if err != nil {
			return nil, err
		}
		return buildCreateVolumeResponse(share.ID, int(size), req, accessibleTopology), nil
	}
	turboResponse, err := services.CreateShareCompleted(cloud, &createShareOpts)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully createVolume response detail: %v", protosanitizer.StripSecrets(turboResponse))
	return buildCreateVolumeResponse(turboResponse.ID, sizeInGiB, req, accessibleTopology), nil
}

// tryGetSubDirContext checks whether the CreateVolume request is for the
// sub-directory mode by reading the PVC annotations. It uses the PVC name
// and namespace injected by csi-provisioner (--extra-create-metadata) to
// fetch the PVC object via the Kubernetes API. Returns (nil, nil) if the
// request does not carry PVC metadata (not a sub-directory request).
func (cs *controllerServer) tryGetSubDirContext(parameters map[string]string) (*SubDirVolumeContext, error) {
	kubeClient := cs.Driver.kubeClient
	if kubeClient == nil {
		// Kubernetes client is unavailable (e.g. RBAC not configured). We
		// cannot read PVC annotations to determine whether this is a
		// sub-directory request. Return (nil, nil) so that the legacy
		// create-new-share path proceeds normally. A PVC that actually
		// targets sub-directory mode will fail later in the legacy path
		// (typically with an AvailabilityZone validation error), which is
		// the expected symptom of a missing kubeClient -- far better than
		// blocking ALL CreateVolume calls including legacy ones.
		log.Warningf("Kubernetes client is not initialized; cannot evaluate " +
			"PVC annotations for sub-directory mode. Falling back to " +
			"legacy create-new-share mode. Ensure the driver's RBAC/" +
			"ServiceAccount has get/list/watch on PVCs to enable " +
			"sub-directory mode.")
		return nil, nil
	}
	pvcName := parameters[PvcNameKey]
	pvcNamespace := parameters[PvcNamespaceKey]
	if pvcName == "" || pvcNamespace == "" {
		// No PVC metadata injected by csi-provisioner; not a sub-directory request.
		return nil, nil
	}
	subDirCtx, err := kubeClient.GetSubDirVolumeContext(pvcName, pvcNamespace)
	if err != nil {
		log.Errorf("Failed to get sub-directory context from PVC %s/%s: %v. "+
			"Refusing to fall back to create-new-share mode to avoid unintended "+
			"volume provisioning.", pvcNamespace, pvcName, err)
		return nil, status.Errorf(codes.Internal,
			"Failed to read PVC %s/%s annotations for sub-directory mode: %v. "+
				"Check the driver's RBAC/ServiceAccount permissions on PVCs.",
			pvcNamespace, pvcName, err)
	}
	return subDirCtx, nil
}

// createSubDirVolume handles the sub-directory provisioning mode: it uses an
// existing SFS Turbo instance, creates (or reuses) a sub-directory, optionally
// sets a quota, and returns a CreateVolumeResponse with an encoded VolumeId.
func (cs *controllerServer) createSubDirVolume(req *csi.CreateVolumeRequest, ctx *SubDirVolumeContext) (
	*csi.CreateVolumeResponse, error) {
	cloud := cs.Driver.cloud

	if err := ValidateSubDirVolumeContext(ctx); err != nil {
		return nil, err
	}

	log.Infof("CreateVolume sub-directory mode: shareID=%s, path=%s, reclaimPolicy=%s, quotaEnabled=%v",
		ctx.ShareID, ctx.Path, ctx.ReclaimPolicy, ctx.QuotaEnabled)

	// Verify the SFS Turbo instance exists. If not, return an error.
	share, err := services.GetShare(cloud, ctx.ShareID)
	if err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound,
				"SFS Turbo instance %s does not exist: %v", ctx.ShareID, err)
		}
		return nil, status.Errorf(codes.Internal,
			"Failed to query SFS Turbo instance %s: %v", ctx.ShareID, err)
	}
	log.Infof("CreateVolume sub-directory mode: found existing share %s (export: %s, size: %s)",
		share.ID, share.ExportLocation, share.Size)

	// Create the sub-directory via the REST API. If it already exists (409),
	// the service layer treats it as success (idempotent).
	if err := services.CreateDir(cloud, ctx.ShareID, ctx.Path); err != nil {
		return nil, err
	}
	log.Infof("CreateVolume sub-directory mode: directory %s is ready on share %s", ctx.Path, ctx.ShareID)

	// Optionally set the directory quota to match the PVC requested storage.
	capacityBytes := req.GetCapacityRange().GetRequiredBytes()
	var reportedCapacity int64
	if ctx.QuotaEnabled {
		capacityMB := fs.PathToCapacityMB(capacityBytes)
		if capacityMB <= 0 {
			return nil, status.Errorf(codes.InvalidArgument,
				"Requested capacity must be positive when quota is enabled, got %d bytes", capacityBytes)
		}
		if err := services.EnsureQuota(cloud, ctx.ShareID, ctx.Path, capacityMB); err != nil {
			return nil, err
		}
		log.Infof("CreateVolume sub-directory mode: quota set to %d MB for path %s on share %s",
			capacityMB, ctx.Path, ctx.ShareID)
		// Report the actual quota capacity (MB rounded up) so Kubernetes
		// records a value consistent with what is enforced on the share.
		reportedCapacity = int64(capacityMB) * 1024 * 1024
	} else {
		// No quota: report the underlying share's actual capacity so the PV
		// capacity is consistent with ControllerGetVolume.
		shareSize, err := strconv.ParseFloat(share.Size, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal,
				"Failed to convert share size %q to number: %v", share.Size, err)
		}
		reportedCapacity = int64(shareSize * float64(common.GbByteSize))
	}

	// Build the VolumeId as "<shareID>#<path>" and the VolumeContext to pass
	// the reclaim policy and quota info to the node and to DeleteVolume.
	volumeID := encodeVolumeID(ctx.ShareID, ctx.Path)
	volumeContext := map[string]string{
		VolumeContextShareID:       ctx.ShareID,
		VolumeContextPath:          ctx.Path,
		VolumeContextReclaimPolicy: ctx.ReclaimPolicy,
		VolumeContextQuotaEnabled:  strconv.FormatBool(ctx.QuotaEnabled),
		VolumeContextIsSubDir:      "true",
	}

	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: reportedCapacity,
			VolumeContext: volumeContext,
		},
	}
	log.Infof("CreateVolume sub-directory mode: successfully created volume %s", volumeID)
	return response, nil
}

func getAutoCreateSGRules(parameters map[string]string) (*string, error) {
	var autoCreateSGRules *string
	if val, ok := parameters[AutoCreateSecurityGroupRules]; ok && val != "" {
		if val != "true" && val != "false" {
			return nil, status.Errorf(codes.InvalidArgument,
				"Validation failed, autoCreateSecurityGroupRules must be \"true\" or \"false\", got: %s", val)
		}
		autoCreateSGRules = &val
	}
	return autoCreateSGRules, nil
}

func (cs *controllerServer) getAccessibleTopology(req *csi.CreateVolumeRequest, volumeAz string) ([]*csi.Topology, error) {
	if len(volumeAz) > 0 {
		return nil, nil
	}

	idc := cs.Driver.cloud.Global.Idc
	if idc {
		return nil, status.Errorf(codes.InvalidArgument, "availability cannot be empty when IDC is true")
	}

	if volumeAz = common.GetAZFromTopology(req.GetAccessibilityRequirements(), topologyKey); volumeAz == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Validation failed, volumeAz cannot be empty")
	}
	log.Infof("Get AZ By GetAccessibilityRequirements Availability Zone: %s", volumeAz)
	// Topology need to be returned by response
	return []*csi.Topology{
		{Segments: map[string]string{topologyKey: volumeAz}},
	}, nil
}

func buildCreateVolumeResponse(shareID string, sizeInGiB int, req *csi.CreateVolumeRequest,
	accessibleTopology []*csi.Topology) *csi.CreateVolumeResponse {
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:           shareID,
			ContentSource:      req.GetVolumeContentSource(),
			CapacityBytes:      int64(sizeInGiB) * common.GbByteSize,
			AccessibleTopology: accessibleTopology,
		},
	}
}

func checkVolumeExists(cloud *config.CloudCredentials, share shares.Share, isHpc bool) (
	*shares.Turbo, error) {
	turbos, err := services.ListTotalShares(cloud)
	if err != nil {
		return nil, err
	}
	log.Infof("CreateVolume checkVolumeExists list total shares: %v", protosanitizer.StripSecrets(turbos))
	for _, v := range turbos {
		if v.Name == share.Name {
			size, err := strconv.ParseFloat(v.Size, 64)
			if err != nil {
				return nil, status.Errorf(codes.Internal,
					"Failed to convert string size to number size, %v", v.Size)
			}
			// For HPC shares the SFS Turbo API does not validate share_type, so
			// the value returned from the server may differ from the request.
			// Use the requested share_type in the comparison to avoid spurious
			// idempotency failures.
			shareType := v.ShareType
			if isHpc {
				shareType = share.ShareType
			}
			turboOpts := shares.Share{
				Name:             v.Name,
				ShareProto:       v.ShareProto,
				ShareType:        shareType,
				Size:             int(size),
				AvailabilityZone: v.AvailabilityZone,
				VpcID:            v.VpcID,
				SubnetID:         v.SubnetID,
				SecurityGroupID:  v.SecurityGroupID,
				Metadata:         share.Metadata,
			}
			if reflect.DeepEqual(turboOpts, share) {
				return &v, nil
			}
			return nil, status.Errorf(codes.InvalidArgument,
				"SFS-Turbo name: %s already exists with different attributes", share.Name)
		}
	}
	return nil, nil
}

func createVolumeValidation(name string, capacityRange *csi.CapacityRange) error {
	if len(name) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, name cannot be empty")
	}

	if capacityRange == nil {
		return status.Error(codes.InvalidArgument, "Validation failed, capacityRange cannot be empty")
	}
	return nil
}

// resolveExpandParams reads the expandType and hpcBw StorageClass parameters
// with loose inference: when hpcBw is set without expandType, expandType is
// inferred as "hpc". When neither is set, both are empty and the request
// targets a legacy STANDARD/PERFORMANCE share.
func resolveExpandParams(parameters map[string]string) (expandType, hpcBw string) {
	expandType = parameters[ExpandType]
	hpcBw = parameters[HpcBw]
	if expandType == "" && hpcBw != "" {
		expandType = expandTypeHpc
	}
	return expandType, hpcBw
}

// validateLegacyCreateSize enforces the original 500~32768 GiB range used by
// non-HPC shares and auto-raises sizes below the minimum. HPC shares bypass
// this and rely on the SFS Turbo API for capacity validation.
func validateLegacyCreateSize(sizeInGiB int) (int, error) {
	if sizeInGiB > maxSizeInGiB {
		return 0, status.Errorf(codes.InvalidArgument,
			"Validation failed, required size %v GB exceeds the max size %v GB", sizeInGiB, maxSizeInGiB)
	}
	if sizeInGiB < minSizeInGiB {
		return minSizeInGiB, nil
	}
	return sizeInGiB, nil
}

func (cs *controllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {
	log.Infof("DeleteVolume called with request %v", protosanitizer.StripSecrets(*req))
	cloud := cs.Driver.cloud

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}

	// Check if this is a sub-directory volume (VolumeId contains "#").
	if isSubDirVolumeID(volumeID) {
		return cs.deleteSubDirVolume(req, volumeID)
	}

	if err := services.DeleteShareCompleted(cloud, volumeID); err != nil {
		return nil, err
	}
	log.Infof("Successfully deleted volume %s", volumeID)
	return &csi.DeleteVolumeResponse{}, nil
}

// deleteSubDirVolume handles deletion for the sub-directory mode. It reads the
// reclaim policy from the PV's volumeAttributes (via the K8s API) and decides
// whether to delete the sub-directory and its quota or retain them.
func (cs *controllerServer) deleteSubDirVolume(req *csi.DeleteVolumeRequest, volumeID string) (
	*csi.DeleteVolumeResponse, error) {
	cloud := cs.Driver.cloud

	shareID, path := decodeVolumeID(volumeID)
	log.Infof("DeleteVolume sub-directory mode: shareID=%s, path=%s", shareID, path)

	// Determine the reclaim policy. Try to read it from the PV's
	// volumeAttributes. If we cannot determine it, default to
	// retain-volume-only for safety.
	reclaimPolicy := ReclaimPolicyRetainVolumeOnly
	if cs.Driver.kubeClient != nil {
		pvName := ""
		if req.GetSecrets() != nil {
			pvName = req.GetSecrets()[PvNameKey]
		}
		if pvName != "" {
			rp, err := cs.Driver.kubeClient.GetReclaimPolicyFromPV(pvName)
			if err != nil {
				log.Warningf("Failed to get reclaim policy from PV %s: %v. Using default %s",
					pvName, err, ReclaimPolicyRetainVolumeOnly)
			} else {
				reclaimPolicy = rp
			}
		} else {
			rp, err := cs.Driver.kubeClient.GetReclaimPolicyByVolumeHandle(volumeID)
			if err != nil {
				log.Warningf("Failed to find PV by volumeHandle %s: %v. Using default %s",
					volumeID, err, ReclaimPolicyRetainVolumeOnly)
			} else {
				reclaimPolicy = rp
			}
		}
	} else {
		log.Infof("Kubernetes client not available, defaulting to %s", reclaimPolicy)
	}

	log.Infof("DeleteVolume sub-directory mode: reclaimPolicy=%s", reclaimPolicy)

	if reclaimPolicy == ReclaimPolicyRetainVolumeOnly {
		log.Infof("DeleteVolume sub-directory mode: retain-volume-only, keeping directory %s on share %s",
			path, shareID)
		return &csi.DeleteVolumeResponse{}, nil
	}

	// reclaimPolicy == ReclaimPolicyDelete: delete the sub-directory and quota.
	if err := services.DeleteQuota(cloud, shareID, path); err != nil {
		return nil, err
	}
	log.Infof("DeleteVolume sub-directory mode: deleted quota for path %s on share %s", path, shareID)

	if err := services.DeleteDir(cloud, shareID, path); err != nil {
		return nil, err
	}
	log.Infof("DeleteVolume sub-directory mode: deleted directory %s on share %s", path, shareID)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerGetVolume(_ context.Context, req *csi.ControllerGetVolumeRequest) (
	*csi.ControllerGetVolumeResponse, error) {
	log.Infof("ControllerGetVolume called with request %v", protosanitizer.StripSecrets(*req))
	cloud := cs.Driver.cloud
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}

	// For sub-directory mode, decode the shareID from the encoded VolumeId.
	shareID, _ := decodeVolumeID(volumeID)

	volume, err := services.GetShare(cloud, shareID)
	if err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "Volume %s not exist: %v", volumeID, err)
		}
		return nil, status.Errorf(codes.Internal, "Failed to query volume %s, error: %v", volumeID, err)
	}
	size, err := strconv.ParseFloat(volume.Size, 64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert string size to number size: %v", volume.Size)
	}

	response := csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: int64(size * common.GbByteSize),
		},
		Status: &csi.ControllerGetVolumeResponse_VolumeStatus{},
	}
	log.Infof("Successfully get volume detail: %v", protosanitizer.StripSecrets(response))
	return &response, nil
}

func (cs *controllerServer) ControllerPublishVolume(_ context.Context, _ *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerUnpublishVolume(_ context.Context, _ *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListVolumes(_ context.Context, req *csi.ListVolumesRequest) (
	*csi.ListVolumesResponse, error) {
	log.Infof("ListVolumes called with request %v", protosanitizer.StripSecrets(*req))
	cloud := cs.Driver.cloud

	if req.GetMaxEntries() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Validation failed, maxEntries cannot be zero")
	}
	offset, err := strconv.Atoi(req.GetStartingToken())
	if err != nil {
		offset = 0
	}
	opts := shares.ListOpts{
		Limit:  int(req.GetMaxEntries()),
		Offset: offset,
	}
	pageList, err := services.ListPageShares(cloud, opts)
	if err != nil {
		return nil, err
	}

	entries := make([]*csi.ListVolumesResponse_Entry, 0, len(pageList.Shares))
	for _, element := range pageList.Shares {
		size, err := strconv.ParseFloat(element.Size, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert string size to number size, %v", element.Size)
		}
		volume := csi.Volume{
			VolumeId:      element.ID,
			CapacityBytes: int64(size * common.GbByteSize),
		}
		entry := csi.ListVolumesResponse_Entry{
			Volume: &volume,
		}
		entries = append(entries, &entry)
	}
	response := &csi.ListVolumesResponse{Entries: entries}
	currentOffset := opts.Offset + len(entries)
	if currentOffset < pageList.Count {
		response.NextToken = strconv.Itoa(currentOffset)
	}
	log.Infof("Successful query volume list. detail: %v", protosanitizer.StripSecrets(response))
	return response, nil
}

func (cs *controllerServer) CreateSnapshot(_ context.Context, _ *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) DeleteSnapshot(_ context.Context, _ *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListSnapshots(_ context.Context, _ *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerGetCapabilities(_ context.Context, req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {
	log.Infof("ControllerGetCapabilities called with request %v", protosanitizer.StripSecrets(*req))
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(_ context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Infof("ValidateVolumeCapabilities called with request %v", protosanitizer.StripSecrets(*req))

	reqVolCap := req.GetVolumeCapabilities()
	if len(reqVolCap) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Validation failed, Volume Capabilities cannot be empty")
	}
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Validation failed, Volume ID cannot be empty")
	}
	cloud := cs.Driver.cloud

	// For sub-directory mode, decode the shareID from the encoded VolumeId.
	shareID, _ := decodeVolumeID(volumeID)

	if _, err := services.GetShare(cloud, shareID); err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound,
				"ValidateVolumeCapabiltites Volume: %s not fount, Error: %v", volumeID, err)
		}
		return nil, status.Errorf(codes.Internal,
			"ValidateVolumeCapabiltites Failed to Get Volume: %s, Error: %v", volumeID, err)
	}

	m := make(map[csi.VolumeCapability_AccessMode_Mode]bool, len(cs.Driver.vcap))
	for _, v := range cs.Driver.vcap {
		m[v.Mode] = true
	}

	var volumeCapabilities []*csi.VolumeCapability
	for _, c := range reqVolCap {
		if m[c.GetAccessMode().GetMode()] {
			volumeCapability := &csi.VolumeCapability{
				AccessMode: &csi.VolumeCapability_AccessMode{
					Mode: c.GetAccessMode().GetMode(),
				},
			}
			volumeCapabilities = append(volumeCapabilities, volumeCapability)
		}
	}
	confirmed := &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: volumeCapabilities}
	return &csi.ValidateVolumeCapabilitiesResponse{Confirmed: confirmed}, nil
}

func (cs *controllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetCapacity is not yet implemented")
}

func (cs *controllerServer) ControllerExpandVolume(_ context.Context, req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error) {
	log.Infof("ControllerExpandVolume called with request %v", protosanitizer.StripSecrets(*req))
	volumeID := req.GetVolumeId()
	capacityRange := req.GetCapacityRange()
	if err := expandVolumeValidation(volumeID, capacityRange); err != nil {
		return nil, err
	}
	sizeInGiB := int(utils.RoundUpSize(capacityRange.GetRequiredBytes(), common.GbByteSize))

	cloud := cs.Driver.cloud

	// Check if this is a sub-directory volume (VolumeId contains "#").
	if isSubDirVolumeID(volumeID) {
		return cs.expandSubDirVolume(req, volumeID, capacityRange)
	}

	volume, err := services.GetShare(cloud, volumeID)
	if err != nil {
		return nil, err
	}

	// HPC shares (created with expand_type=hpc) skip the legacy max-size and
	// 100GiB minimum-step checks; the SFS Turbo API validates the target size.
	isHpc := volume.ExpandType == expandTypeHpc
	if !isHpc && sizeInGiB > maxSizeInGiB {
		return nil, status.Errorf(codes.OutOfRange,
			"Validation failed, expand required size %v exceeds the max size %v", sizeInGiB, maxSizeInGiB)
	}

	// current volume size
	currentSizeInGiBFloat, err := strconv.ParseFloat(volume.Size, 64)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert string size to number size, %s", volume.Size)
	}
	currentSizeInGiB := int(currentSizeInGiBFloat)
	if currentSizeInGiB >= sizeInGiB {
		log.Warningf("Volume %v has been already expanded to %v, requested %v", volumeID, currentSizeInGiB, sizeInGiB)
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(currentSizeInGiB * common.GbByteSize),
			NodeExpansionRequired: false,
		}, nil
	}

	// HPC shares leave the expand target to the API; legacy shares enforce the
	// 100GiB minimum expand step.
	if !isHpc && sizeInGiB-currentSizeInGiB < expandStepInGiB {
		// reset sizeInGiB
		sizeInGiB = currentSizeInGiB + expandStepInGiB
		if sizeInGiB > maxSizeInGiB {
			return nil, status.Errorf(codes.OutOfRange,
				"Validation failed, required step size less than 100G; expand required size %v exceeds the max size %v",
				sizeInGiB, maxSizeInGiB)
		}
	}

	if err = services.ExpandShareCompleted(cloud, volumeID, sizeInGiB); err != nil {
		return nil, err
	}
	log.Infof("Successfully resized volume %v to size %v", volumeID, sizeInGiB)
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(sizeInGiB) * common.GbByteSize,
		NodeExpansionRequired: false,
	}, nil
}

// expandSubDirVolume handles the expand for the sub-directory mode. Since the
// underlying SFS Turbo instance is shared, we cannot expand it for a single
// sub-directory. Instead, if quota is enabled, we update the directory quota
// to match the new requested capacity.
func (cs *controllerServer) expandSubDirVolume(req *csi.ControllerExpandVolumeRequest,
	volumeID string, capacityRange *csi.CapacityRange) (
	*csi.ControllerExpandVolumeResponse, error) {
	cloud := cs.Driver.cloud

	shareID, path := decodeVolumeID(volumeID)
	log.Infof("ControllerExpandVolume sub-directory mode: shareID=%s, path=%s", shareID, path)

	// Verify the share still exists.
	share, err := services.GetShare(cloud, shareID)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Failed to query SFS Turbo instance %s: %v", shareID, err)
	}

	// Check if quota is enabled by looking at the existing quota.
	quota, exists, err := services.ShowQuota(cloud, shareID, path)
	if err != nil {
		return nil, err
	}

	if !exists {
		// No quota set. Expand is a no-op; return the share's current size.
		log.Infof("ControllerExpandVolume sub-directory mode: no quota on path %s, expand is no-op", path)
		shareSize, _ := strconv.ParseFloat(share.Size, 64)
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(shareSize * float64(common.GbByteSize)),
			NodeExpansionRequired: false,
		}, nil
	}

	// Update the quota to match the new requested capacity.
	capacityMB := fs.PathToCapacityMB(capacityRange.GetRequiredBytes())
	if capacityMB <= 0 {
		return nil, status.Errorf(codes.InvalidArgument,
			"Requested capacity must be positive, got %d bytes", capacityRange.GetRequiredBytes())
	}

	// Only grow the quota, never shrink it.
	if quota.Capacity >= capacityMB {
		log.Infof("ControllerExpandVolume sub-directory mode: existing quota %d MB >= requested %d MB, no update needed",
			quota.Capacity, capacityMB)
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(quota.Capacity) * 1024 * 1024,
			NodeExpansionRequired: false,
		}, nil
	}

	if err := services.UpdateQuota(cloud, shareID, path, capacityMB); err != nil {
		return nil, err
	}
	log.Infof("ControllerExpandVolume sub-directory mode: updated quota from %d MB to %d MB for path %s",
		quota.Capacity, capacityMB, path)

	// Report the actual quota capacity (MB rounded up) so Kubernetes records
	// a value consistent with what is enforced on the share.
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(capacityMB) * 1024 * 1024,
		NodeExpansionRequired: false,
	}, nil
}

func expandVolumeValidation(volumeID string, capacityRange *csi.CapacityRange) error {
	if len(volumeID) == 0 {
		return status.Errorf(codes.InvalidArgument, "Validation failed, volume Id cannot be empty")
	}

	if capacityRange == nil {
		return status.Errorf(codes.InvalidArgument, "Validation failed, capacity range cannot be nil")
	}
	return nil
}
