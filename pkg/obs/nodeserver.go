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
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs/services"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/metadatas"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/mounts"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
	utilpath "k8s.io/utils/path"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type nodeServer struct {
	Driver      *Driver
	Mount       mounts.IMount
	Metadata    metadatas.IMetadata
	MountClient http.Client
}

const (
	credentialFile = "/dev/csi-tool/passwd-obsfs"
	perm           = 0600
)

func (ns *nodeServer) NodeStageVolume(_ context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	log.Infof("NodeStageVolume: called with args %v", protosanitizer.StripSecrets(*req))
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeUnstageVolume(_ context.Context, req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {
	log.Infof("NodeUnstageVolume: called with args %v", protosanitizer.StripSecrets(*req))
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

	credentials := ns.Driver.cloud
	volume, err := services.GetParallelFSBucket(credentials, volumeID)
	if err != nil {
		return nil, err
	}
	log.Infof("NodePublishVolume: volume detail: %s", protosanitizer.StripSecrets(volume))

	notMnt, err := ns.Mount.IsLikelyNotMountPointAttach(targetPath)
	if err != nil {
		return nil, err
	}
	if !notMnt {
		log.Infof("NodePublishVolume: %s has already mounted", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}
	if err := ns.Mount.MakeDir(targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to make dir: %s, error: %v", targetPath, err)
	}
	if err := createCredentialFile(ns.Driver.cloud.Global.AccessKey, ns.Driver.cloud.Global.SecretKey); err != nil {
		return nil, err
	}

	parameters := map[string]string{
		"bucketName": volume.BucketName,
		"targetPath": targetPath,
		"region":     ns.Driver.cloud.Global.Region,
		"cloud":      ns.Driver.cloud.Global.Cloud,
		"credential": credentialFile,
	}
	md5String, err := utils.Md5SortMap(parameters)
	if err != nil {
		return nil, err
	}
	token, err := utils.EncryptAESCBC(Secret, md5String)
	if err != nil {
		return nil, err
	}
	commandRPC := CommandRPC{
		Action:     ActionMount,
		Token:      token,
		Parameters: parameters,
	}
	if err := sendCommand(commandRPC, ns.MountClient); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to mount %s at %s: %v",
			volume.BucketName, targetPath, err)
	}
	log.Infof("NodePublishVolume: mount %s at %s successfully", volume.BucketName, targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

func createCredentialFile(accessKey, secretKey string) error {
	if err := os.MkdirAll(path.Dir(credentialFile), os.ModePerm); err != nil {
		return status.Errorf(codes.Internal, "Failed to make dir: %s, error: %v", credentialFile, err)
	}

	writer, err := os.OpenFile(filepath.Clean(credentialFile), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to open file: %s, error: %v", credentialFile, err)
	}
	defer writer.Close()

	credentialInfo := accessKey + ":" + secretKey
	_, err = fmt.Fprintln(writer, credentialInfo)
	return err
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

	credentials := ns.Driver.cloud
	volume, err := services.GetParallelFSBucket(credentials, volumeID)
	if err != nil {
		return nil, err
	}
	log.Infof("NodeUnpublishVolume: volume detail: %s", protosanitizer.StripSecrets(volume))
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
	log.Infof("NodeGetInfo called with request %v", protosanitizer.StripSecrets(*req))

	nodeID, err := ns.Metadata.GetInstanceID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve instance id of node %s", err)
	}

	return &csi.NodeGetInfoResponse{
		NodeId: nodeID,
	}, nil
}

func (ns *nodeServer) NodeGetCapabilities(_ context.Context, req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {
	log.Infof("NodeGetCapabilities called with req: %#v", protosanitizer.StripSecrets(*req))

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
	capacity, usedBytes := stats.TotalBytes, stats.UsedBytes

	bucket, err := services.GetParallelFSBucket(ns.Driver.cloud, volumeID)
	if err != nil {
		return nil, err
	}
	if bucket.Capacity != 0 {
		capacity = bucket.Capacity
	}
	used, _, err := services.GetBucketStorage(ns.Driver.cloud, volumeID)
	if err != nil {
		return nil, err
	}
	if used != 0 {
		usedBytes = used
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{Total: capacity, Available: capacity - usedBytes, Used: usedBytes, Unit: csi.VolumeUsage_BYTES},
			{Total: stats.TotalInodes, Available: stats.AvailableInodes, Used: stats.UsedInodes, Unit: csi.VolumeUsage_INODES},
		},
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(_ context.Context, _ *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
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
