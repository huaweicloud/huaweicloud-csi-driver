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
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/chnsz/golangsdk"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

type nodeServer struct {
	Driver   *SfsTurboDriver
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(2).Infof("NodePublishVolume called with request %v", *req)
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	target := req.GetTargetPath()
	if len(target) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	//Get Volume
	volumeID := req.GetVolumeId()
	client, err := ns.Driver.cloud.SFSTurboV1Client()
    if err != nil {
		klog.V(3).Infof("NodePublishVolume Failed to create SFS v2 client: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
    }

	// wait create new share
	err = waitForShareStatus(client, volumeID)
	if err != nil {
		return nil, err
	}

	share, err := getShare(client, volumeID)
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("NodePublishVolume Volume %s not found", volumeID))
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("NodePublishVolume %v", err))
	}

	//Get Volume export location
	source := share.ExportLocation
	if source == "" {
		return nil, status.Error(codes.Internal, fmt.Sprintf("NodePublishVolume Volume %s location not found", volumeID))
	}

	mountOptions := "noresvport,nolock"
	if req.GetReadonly() {
		mountOptions += ",ro"
	}

	if isMounted(target) {
		klog.V(2).Infof("NodePublishVolume: %s is already mounted", target)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	klog.V(2).Infof("NodePublishVolume: creating dir %s", target)
	if err := makeDir(target); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not create dir %q: %v", target, err)
	}

	klog.V(2).Infof("NodePublishVolume: mounting %s at %s with mountOptions: %v", source, target, mountOptions)
	if err := Mount(source, target, mountOptions); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return nil, status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return nil, status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}
	klog.V(2).Infof("NodePublishVolume: mount %s at %s successfully", source, target)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.V(2).Infof("NodeUnPublishVolume: called with args %+v", *req)
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	klog.V(2).Infof("NodeUnpublishVolume: unmounting volume %s on %s", volumeID, targetPath)
	err := Unmount(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmount target %q: %v", targetPath, err)
	}
	klog.V(2).Infof("NodeUnpublishVolume: unmount volume %s on %s successfully", volumeID, targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.V(2).Infof("NodeGetInfo called with request %v", *req)
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.V(2).Infof("NodeGetCapabilities called with req: %#v", req)

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: ns.Driver.nscap,
	}, nil
}

func (ns *nodeServer) NodeGetVolumeStats(_ context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeExpandVolume node expand volume
func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
