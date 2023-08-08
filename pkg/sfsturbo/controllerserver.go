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
)

// resourcemode from SC
const (
	Availability = "availability"
	ShareType    = "shareType"
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
	sizeInGiB := int(utils.RoundUpSize(capacityRange.GetRequiredBytes(), common.GbByteSize))
	// 500 ~ 32768
	if sizeInGiB > maxSizeInGiB {
		return nil, status.Errorf(codes.InvalidArgument,
			"Validation failed, required size %v GB exceeds the max size %v GB", sizeInGiB, maxSizeInGiB)
	}
	if sizeInGiB < minSizeInGiB {
		sizeInGiB = minSizeInGiB
	}

	var accessibleTopology []*csi.Topology
	parameters := req.GetParameters()
	// First check if volAvailability is already specified, if not get preferred from Topology
	// Required, incase vol AZ is different from node AZ
	volumeAz := parameters[Availability]
	if len(volumeAz) == 0 {
		if volumeAz = common.GetAZFromTopology(req.GetAccessibilityRequirements(), topologyKey); volumeAz == "" {
			return nil, status.Errorf(codes.InvalidArgument, "Validation failed, volumeAz cannot be empty")
		}
		log.Infof("Get AZ By GetAccessibilityRequirements Availability Zone: %s", volumeAz)
		// Topology need to be returned by response
		accessibleTopology = []*csi.Topology{
			{Segments: map[string]string{topologyKey: volumeAz}},
		}
	}

	shareType := parameters[ShareType]
	if len(shareType) == 0 {
		shareType = defaultShareType
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
	}
	log.Infof("CreateVolume creating param: %v", protosanitizer.StripSecrets(createShareOpts))
	// Check if there are any volumes with the same name
	share, err := checkVolumeExists(cloud, createShareOpts)
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

func checkVolumeExists(cloud *config.CloudCredentials, share shares.Share) (
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
			turboOpts := shares.Share{
				Name:             v.Name,
				ShareProto:       v.ShareProto,
				ShareType:        v.ShareType,
				Size:             int(size),
				AvailabilityZone: v.AvailabilityZone,
				VpcID:            v.VpcID,
				SubnetID:         v.SubnetID,
				SecurityGroupID:  v.SecurityGroupID,
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

func (cs *controllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (
	*csi.DeleteVolumeResponse, error) {
	log.Infof("DeleteVolume called with request %v", protosanitizer.StripSecrets(*req))
	cloud := cs.Driver.cloud

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}

	if err := services.DeleteShareCompleted(cloud, volumeID); err != nil {
		return nil, err
	}
	log.Infof("Successfully deleted volume %s", volumeID)
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
	volume, err := services.GetShare(cloud, volumeID)
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

	if _, err := services.GetShare(cloud, volumeID); err != nil {
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
	if sizeInGiB > maxSizeInGiB {
		return nil, status.Errorf(codes.OutOfRange,
			"Validation failed, expand required size %v exceeds the max size %v", sizeInGiB, maxSizeInGiB)
	}

	cloud := cs.Driver.cloud
	volume, err := services.GetShare(cloud, volumeID)
	if err != nil {
		return nil, err
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

	if sizeInGiB-currentSizeInGiB < expandStepInGiB {
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
		CapacityBytes:         int64(sizeInGiB),
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
