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

package sfs

import (
	"fmt"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/sfs/v2/shares"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"k8s.io/klog"
)

type controllerServer struct {
	Driver *SfsDriver
}

func (cs *controllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {
	log.Infof("CreateVolume called with request %v", protosanitizer.StripSecrets(*req))
	client, err := cs.Driver.cloud.SFSV2Client()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create SFS v2 client: %v", err)
	}
	name := req.GetName()
	capacityRange := req.GetCapacityRange()
	if err := createVolumeValidation(name, capacityRange); err != nil {
		return nil, err
	}

	requestedSize := capacityRange.GetRequiredBytes()
	if requestedSize == 0 {
		// At least 10GiB
		requestedSize = 10 * common.GbByteSize
	}
	sizeInGiB := int(utils.RoundUpSize(requestedSize, common.GbByteSize))
	share, err := checkVolumeExist(client, name, sizeInGiB, cs.Driver.shareProto)
	if err != nil {
		return nil, err
	}
	if share == nil {
		// Creating a share
		createOpts := shares.CreateOpts{
			ShareProto: cs.Driver.shareProto,
			Size:       sizeInGiB,
			Name:       req.GetName(),
		}
		share, err = createShare(client, &createOpts)
		if err != nil {
			return nil, err
		}
	}

	// Grant access to the share
	log.Infof("Creating an access rule to VPC %s", cs.Driver.cloud.Vpc.ID)
	if err := grantAccess(client, share.ID, cs.Driver.cloud.Vpc.ID); err != nil {
		return nil, err
	}
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      share.ID,
			ContentSource: req.GetVolumeContentSource(),
			CapacityBytes: int64(sizeInGiB) * common.GbByteSize,
		},
	}, nil
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

func checkVolumeExist(client *golangsdk.ServiceClient, name string, sizeInGiB int, shareProto string) (
	*shares.Share, error) {
	opts := shares.ListOpts{
		Name: name,
	}
	list, err := shareList(client, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Failed to query SFS by name, cannot verify whether it exists: %v", err)
	}
	for _, v := range list {
		if v.Name == name {
			// if name, share proto and size are the same, return success
			if v.ShareProto == shareProto && v.Size == sizeInGiB {
				return &v, nil
			}
			return nil, status.Errorf(codes.InvalidArgument,
				"SFS name: %s already exists with a different size or share_proto", name)
		}
	}
	return nil, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.V(2).Infof("DeleteVolume called with request %v", *req)

	volID := req.GetVolumeId()
	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	client, err := cs.Driver.cloud.SFSV2Client()
	if err != nil {
		klog.V(3).Infof("Failed to create SFS v2 client: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = deleteShare(client, volID)
	if err != nil {
		klog.V(3).Infof("Failed to DeleteVolume: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteVolume failed with error %v", err))
	}

	klog.V(4).Infof("Delete volume %s", volID)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities implements the default GRPC callout.
// Default supports all capabilities
func (cs *controllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.V(2).Infof("ControllerGetCapabilities called with request %v", *req)

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	klog.V(2).Infof("ValidateVolumeCapabilities called with args %+v", *req)

	reqVolCap := req.GetVolumeCapabilities()

	if len(reqVolCap) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ValidateVolumeCapabilities Volume Capabilities must be provided")
	}
	volumeID := req.GetVolumeId()

	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ValidateVolumeCapabilities Volume ID must be provided")
	}

	client, err := cs.Driver.cloud.SFSV2Client()
	if err != nil {
		klog.V(3).Infof("ValidateVolumeCapabilities Failed to create SFS v2 client: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = getShare(client, volumeID)
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("ValidateVolumeCapabiltites Volume %s not found", volumeID))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("ValidateVolumeCapabiltites %v", err))
	}

	for _, c := range reqVolCap {
		if c.GetAccessMode().Mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER {
			return &csi.ValidateVolumeCapabilitiesResponse{}, nil
		}
	}

	confirmed := &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: reqVolCap}
	return &csi.ValidateVolumeCapabilitiesResponse{Confirmed: confirmed}, nil
}

func (cs *controllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCapacity is not yet implemented")
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

/*
func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	klog.V(4).Infof("ControllerExpandVolume: called with args %+v", *req)

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	cap := req.GetCapacityRange()
	if cap == nil {
		return nil, status.Error(codes.InvalidArgument, "Capacity range not provided")
	}

	volSizeBytes := int64(req.GetCapacityRange().GetRequiredBytes())
	volSizeGB := int(RoundUpSize(volSizeBytes, 1024*1024*1024))
	maxVolSize := cap.GetLimitBytes()

	if maxVolSize > 0 && maxVolSize < volSizeBytes {
		return nil, status.Error(codes.OutOfRange, "After round-up, volume size exceeds the limit specified")
	}

	client, err := cs.Driver.cloud.SFSV2Client()
    if err != nil {
		klog.V(3).Infof("Failed to create SFS v2 client: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
    }

	_, err = getShare(client, volumeID)
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			return nil, status.Error(codes.NotFound, "Volume not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("GetVolume failed with error %v", err))
	}

	err = expandShare(client, volumeID, volSizeGB)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Could not resize volume %q to size %v: %v", volumeID, volSizeGB, err))
	}

	klog.V(4).Infof("ControllerExpandVolume resized volume %v to size %v", volumeID, volSizeGB)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         volSizeBytes,
		NodeExpansionRequired: true,
	}, nil
}
*/
