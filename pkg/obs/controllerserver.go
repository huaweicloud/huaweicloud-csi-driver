/*
Copyright 2022 The Kubernetes Authors.

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

package obs

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs/services"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

type controllerServer struct {
	Driver *Driver
}

func (cs *controllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (
	*csi.CreateVolumeResponse, error) {
	log.Infof("CreateVolume: called with args %v", protosanitizer.StripSecrets(*req))
	credentials := cs.Driver.cloud

	volName := req.GetName()
	if err := createVolumeValidation(volName, req.GetVolumeCapabilities()); err != nil {
		return nil, err
	}

	if volume, err := services.GetParallelFSBucket(credentials, volName); err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	} else if volume != nil {
		log.Infof("Volume %s existence, skip creating", volName)
		return buildCreateVolumeResponse(volume), nil
	}

	parameters := req.GetParameters()
	acl := obs.AclType(parameters["acl"])
	if err := services.CreateBucket(credentials, volName, acl); err != nil {
		return nil, err
	}
	tags := []obs.Tag{{
		Key:   "csi",
		Value: "csi-created",
	}}
	if err := services.AddBucketTags(credentials, volName, tags); err != nil {
		return nil, err
	}

	volume, err := services.GetParallelFSBucket(credentials, volName)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully created volume %s", volName)
	return buildCreateVolumeResponse(volume), nil
}

func (cs *controllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {
	log.Infof("DeleteVolume: called with args %v", protosanitizer.StripSecrets(*req))

	volName := req.GetVolumeId()
	if len(volName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}

	credentials := cs.Driver.cloud
	if tags, err := services.ListBucketTags(credentials, volName); err != nil {
		return nil, err
	} else if !isCreatByCSi(tags, "csi") {
		log.Infof("Volume %s does not create by csi, skip deleting", volName)
		return &csi.DeleteVolumeResponse{}, nil
	}

	if err := services.CleanBucket(credentials, volName); err != nil {
		log.Infof("Successfully deleted volume %s", volName)
		return nil, err
	}
	if err := services.DeleteBucket(credentials, volName); err != nil {
		if status.Code(err) == codes.NotFound {
			log.Infof("Volume %s does not exist, skip deleting", volName)
			return &csi.DeleteVolumeResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "Error deleting volume: %s", err)
	}
	log.Infof("Successfully deleted volume %s", volName)

	return &csi.DeleteVolumeResponse{}, nil
}

func isCreatByCSi(tags []obs.Tag, marker string) bool {
	if tags == nil {
		return false
	}
	for _, tag := range tags {
		if tag.Key == marker {
			return true
		}
	}
	return false
}

func (cs *controllerServer) ControllerGetVolume(_ context.Context, req *csi.ControllerGetVolumeRequest) (
	*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
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
	return nil, status.Error(codes.Unimplemented, "")
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
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ValidateVolumeCapabilities(_ context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerExpandVolume(_ context.Context, req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func createVolumeValidation(volumeName string, capabilities []*csi.VolumeCapability) error {
	if len(volumeName) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, volume name cannot be empty")
	}
	if len(capabilities) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, volume capabilities cannot be empty")
	}
	return nil
}

func buildCreateVolumeResponse(vol *services.Bucket) *csi.CreateVolumeResponse {
	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      vol.BucketName,
			CapacityBytes: vol.Capacity,
		},
	}
	return response
}
