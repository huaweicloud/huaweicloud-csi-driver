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
	"fmt"
	"net"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
	utilpath "k8s.io/utils/path"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/sfsturbo/services"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/metadatas"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/mounts"
)

type nodeServer struct {
	Driver   *SfsTurboDriver
	Mount    mounts.IMount
	Metadata metadatas.IMetadata
}

func (ns *nodeServer) NodeStageVolume(_ context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	log.Infof("NodeStageVolume: called with args %v", protosanitizer.StripSecrets(*req))
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeUnstageVolume(_ context.Context, req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {
	log.Infof("NodeStageVolume: called with args %v", protosanitizer.StripSecrets(*req))
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodePublishVolume(_ context.Context, req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {
	log.Infof("NodePublishVolume: called with args %v", protosanitizer.StripSecrets(*req))
	capability := req.GetVolumeCapability()
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	if err := nodePublishValidation(capability, volumeID, targetPath); err != nil {
		return nil, err
	}

	cloud := ns.Driver.cloud
	share, err := services.GetShare(cloud, volumeID)
	if err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "Share %s has already been deleted.", volumeID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to query share: %s, error: %v", volumeID, err)
	}

	//Get Volume export location
	exportLocation := share.ExportLocation
	if len(exportLocation) == 0 {
		return nil, status.Errorf(codes.Internal, "Not found export location from volume %s", volumeID)
	}

	if isMounted(targetPath) {
		log.Infof("NodePublishVolume: %s has already mounted", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	if err := makeDir(targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to make dir: %s, error: %v", targetPath, err)
	}

	mountOptions := []string{"vers=3,timeo=600,noresvport,nolock"}
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	} else {
		mountOptions = append(mountOptions, "rw")
	}

	log.Infof("NodePublishVolume: mounting %s at %s with mountOptions: %v",
		exportLocation, targetPath, mountOptions)
	if err := ns.Mount.Mounter().Mount(exportLocation, targetPath, "nfs", mountOptions); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to mount %s at %s: %v",
			exportLocation, targetPath, err)
	}
	log.Infof("NodePublishVolume: mount %s at %s successfully", exportLocation, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

func nodePublishValidation(capability *csi.VolumeCapability, volumeID string, targetPath string) error {
	if capability == nil {
		return status.Error(codes.InvalidArgument, "Validation failed, volume capability cannot be nil")
	}
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, volumeID cannot be empty")
	}
	if len(targetPath) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, targetPath cannot be empty")
	}
	return nil
}

func (ns *nodeServer) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {
	log.Infof("NodeUnpublishVolume: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volumeId cannot be empty")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, targetPath cannot be empty")
	}
	log.Infof("NodeUnpublishVolume: unmounting volume %s on %s", volumeID, targetPath)

	notMnt, err := ns.Mount.IsLikelyNotMountPointAttach(targetPath)
	if err != nil {
		return nil, err
	}

	if notMnt {
		log.Infof("NodeUnpublishVolume: %s has already uMounted", targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	if err := ns.Mount.UnmountPath(targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to unmount target %q: %v", targetPath, err)
	}
	log.Infof("NodeUnpublishVolume: unmount volume %s on %s successfully", volumeID, targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetInfo(_ context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Infof("NodeGetInfo: called with args %v", protosanitizer.StripSecrets(*req))

	idc := ns.Driver.cloud.Global.Idc
	if idc {
		log.Infof("IDC is %v, volume will be mounted directly \n", idc)
		macAddress, err := getMACAddress()
		if err != nil {
			log.Errorf("failed to get mac address: %v", err)
			return &csi.NodeGetInfoResponse{}, fmt.Errorf("failed to gen nodeID: %v", err)
		}

		return &csi.NodeGetInfoResponse{
			NodeId: macAddress,
		}, nil
	}

	nodeID, err := ns.Metadata.GetInstanceID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve instance id of node %s", err)
	}

	zone, err := ns.Metadata.GetAvailabilityZone()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve availability zone of node %v", err)
	}
	topology := &csi.Topology{Segments: map[string]string{topologyKey: zone}}
	log.Infof("NodeGetInfo nodeID: %s, topology: %s", nodeID, protosanitizer.StripSecrets(*topology))
	return &csi.NodeGetInfoResponse{
		NodeId:             nodeID,
		AccessibleTopology: topology,
	}, nil
}

func (ns *nodeServer) NodeGetCapabilities(_ context.Context, req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {
	log.Infof("NodeGetCapabilities: called with args %v", protosanitizer.StripSecrets(*req))
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *nodeServer) NodeGetVolumeStats(_ context.Context, req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {
	log.Infof("NodeGetVolumeStats: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	volumePath := req.GetVolumePath()
	if err := nodeGetStatsValidation(volumeID, volumePath); err != nil {
		return nil, err
	}

	stats, err := ns.Mount.GetDeviceStats(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to get stats by path: %s", err)
	}
	log.Infof("NodeGetVolumeStats: stats info :%s", protosanitizer.StripSecrets(*stats))

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Total:     stats.TotalBytes,
				Available: stats.AvailableBytes,
				Used:      stats.UsedBytes,
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Total:     stats.TotalInodes,
				Available: stats.AvailableInodes,
				Used:      stats.UsedInodes,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

func nodeGetStatsValidation(volumeID, volumePath string) error {
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, VolumeID cannot be empty")
	}
	if len(volumePath) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, VolumePath cannot be empty")
	}

	exists, err := utilpath.Exists(utilpath.CheckFollowSymlink, volumePath)
	if err != nil {
		return status.Errorf(codes.Unknown,
			"Failed to check whether VolumePath %s exists: %s", volumePath, err)
	}
	if !exists {
		return status.Errorf(codes.Unknown, "Error, the volume path %s not found", volumePath)
	}
	return nil
}

// NodeExpandVolume node expand volume
func (ns *nodeServer) NodeExpandVolume(_ context.Context, _ *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// getMACAddress will use MAC address as node id if idc is true
func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, v := range interfaces {
		if v.Flags&net.FlagLoopback == 0 && len(v.HardwareAddr.String()) > 0 {
			return strings.ReplaceAll(v.HardwareAddr.String(), ":", "_"), nil
		}
	}

	return "", fmt.Errorf("MAC address not found")
}
