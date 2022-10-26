package evs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chnsz/golangsdk/openstack/evs/v2/cloudvolumes"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
	mountutils "k8s.io/mount-utils"
	utilpath "k8s.io/utils/path"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs/services"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/metadatas"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/mounts"
)

const (
	maxVolumes = 24 // the maximum number of volumes for a KVM instance is 24
)

type nodeServer struct {
	Driver   *EvsDriver
	Mount    mounts.IMount
	Metadata metadatas.IMetadata
}

func (ns *nodeServer) NodeStageVolume(_ context.Context, req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {
	log.Infof("NodeStageVolume: called with args %v", protosanitizer.StripSecrets(*req))

	cc := ns.Driver.cloudCredentials
	stagingTarget := req.GetStagingTargetPath()
	volumeCapability := req.GetVolumeCapability()
	volumeID := req.GetVolumeId()

	vol, err := nodeStageValidation(cc, volumeID, stagingTarget, volumeCapability)
	if err != nil {
		return nil, err
	}

	// Verify whether mounted
	mount := ns.Mount
	// Do not trust the path provided by EVS, get the real path on node
	devicePath, err := getDevicePath(cc, volumeID, mount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to find devicePath for volume: %v", err)
	}

	if blk := volumeCapability.GetBlock(); blk != nil {
		// If block volume, do nothing
		log.Infof("Volume mode is Block, skip staging volume")
		return &csi.NodeStageVolumeResponse{}, nil
	}

	notMnt, err := mount.IsLikelyNotMountPointAttach(stagingTarget)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Volume Mount
	if notMnt {
		// set default fstype is ext4
		fsType := "ext4"
		var options []string
		if mnt := volumeCapability.GetMount(); mnt != nil {
			if mnt.FsType != "" {
				fsType = mnt.FsType
			}
			mountFlags := mnt.GetMountFlags()
			options = append(options, collectMountOptions(fsType, mountFlags)...)
		}
		// Mount volume
		err = mount.Mounter().FormatAndMount(devicePath, stagingTarget, fsType, options)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Try expanding the volume if it's created from a snapshot or another volume (see #1539)
	if vol.SourceVolID != "" || vol.SnapshotID != "" {
		r := mountutils.NewResizeFs(mount.Mounter().Exec)
		needResize, err := r.NeedResize(devicePath, stagingTarget)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "Could not determine if volume %v need to be resized: %v",
				volumeID, err)
		}
		if needResize {
			log.Infof("NodeStageVolume: Resizing volume %v created from a snapshot/volume", volumeID)
			if _, err := r.Resize(devicePath, stagingTarget); err != nil {
				return nil, status.Errorf(codes.Unknown, "Could not resize volume %v:  %v", volumeID, err)
			}
		}
	}

	log.Infof("Successfully staged volume %v", volumeID)
	return &csi.NodeStageVolumeResponse{}, nil
}

func getDevicePath(cc *config.CloudCredentials, volumeID string, mount mounts.IMount) (string, error) {
	volume, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return "", err
	}
	// device type: SCSI
	if volume.Metadata.HwPassthrough == "true" {
		volumeID = volume.WWN
	}

	var devicePath string
	devicePath, _ = mount.GetDevicePath(volumeID)
	if devicePath == "" {
		// try to get from metadata service
		devicePath = metadatas.GetDevicePath(volumeID)
	}

	if len(strings.TrimSpace(devicePath)) == 0 {
		return "", fmt.Errorf("the \"devicePath\" is still empty after getting from mount and metadata.")
	}
	return devicePath, nil
}

func nodeStageValidation(cc *config.CloudCredentials, volumeID, target string, vc *csi.VolumeCapability) (
	*cloudvolumes.Volume, error) {
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, VolumeID cannot be empty")
	}
	if len(target) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, StagingTargetPath cannot be empty")
	}
	if vc == nil {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, VolumeCapability cannot be empty")
	}

	vol, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return nil, err
	}
	return vol, nil
}

func (ns *nodeServer) NodeUnstageVolume(_ context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.
	NodeUnstageVolumeResponse, error) {
	log.Infof("NodeUnstageVolume: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	if err := unstagetValidation(ns.Driver.cloudCredentials, volumeID, stagingTargetPath); err != nil {
		return nil, err
	}

	err := ns.Mount.UnmountPath(stagingTargetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unmount of targetPath %s failed with error %v",
			stagingTargetPath, err)
	}

	log.Infof("Successfully unstaged volume %v at path %s", volumeID, stagingTargetPath)
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func unstagetValidation(cc *config.CloudCredentials, volumeID, target string) error {
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, VolumeID cannot be empty")
	}
	if len(target) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, StagingTargetPath cannot be empty")
	}

	_, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return err
	}
	return nil
}

func (ns *nodeServer) NodePublishVolume(_ context.Context, req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {
	log.Infof("NodePublishVolume: called with args %v", protosanitizer.StripSecrets(*req))

	cc := ns.Driver.cloudCredentials
	volumeID := req.GetVolumeId()
	source := req.GetStagingTargetPath()
	targetPath := req.GetTargetPath()
	volumeCapability := req.GetVolumeCapability()

	if err := nodePublishValidation(cc, volumeID, source, targetPath, volumeCapability); err != nil {
		return nil, err
	}

	ephemeralVolume := req.GetVolumeContext()["csi.storage.k8s.io/ephemeral"] == "true"
	if ephemeralVolume {
		return nodePublishEphemeral(req, ns)
	}

	mountOptions := []string{"bind"}
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	} else {
		mountOptions = append(mountOptions, "rw")
	}

	if blk := volumeCapability.GetBlock(); blk != nil {
		log.Infof("The volume mode is block")
		return nodePublishVolumeForBlock(req, ns, mountOptions)
	}

	mount := ns.Mount
	// Verify whether mounted
	notMnt, err := mount.IsLikelyNotMountPointAttach(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Volume Mount
	if notMnt {
		fsType := "ext4"
		if mnt := volumeCapability.GetMount(); mnt != nil {
			if mnt.FsType != "" {
				fsType = mnt.FsType
			}
		}
		// Mount
		err = mount.Mounter().Mount(source, targetPath, fsType, mountOptions)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	log.Infof("Successfully publish volume on node, targetPath: %s", targetPath)
	return &csi.NodePublishVolumeResponse{}, nil
}

func nodePublishValidation(cc *config.CloudCredentials, volumeID, sourcePath, targetPath string,
	vc *csi.VolumeCapability) error {
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, volumeID cannot be empty")
	}
	if len(targetPath) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, targetPath cannot be empty")
	}
	if vc == nil {
		return status.Error(codes.InvalidArgument, "Validation failed, volumeCapability cannot be empty")
	}
	if len(sourcePath) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, stagingTargetPath cannot be empty")
	}

	_, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return err
	}
	return nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {
	log.Infof("NodeUnpublishVolume: called with args %v", protosanitizer.StripSecrets(*req))

	cc := ns.Driver.cloudCredentials
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, TargetPath cannot be empty")
	}
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, VolumeID cannot be empty")
	}

	ephemeralVolume := false
	vol, err := services.GetVolume(cc, volumeID)
	if err != nil {
		if !common.IsNotFound(err) {
			return nil, status.Errorf(codes.Internal, "Error querying volume details: %s", err)
		}
		// if not found by id, try to search by name
		volName := fmt.Sprintf("ephemeral-%s", volumeID)
		vols, err := services.ListVolumes(cc, cloudvolumes.ListOpts{
			Name: volName,
		})
		//if volume not found then GetVolumesByName returns empty list
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Error querying volume: %v", err)
		}
		if len(vols) > 0 {
			vol = &vols[0]
			ephemeralVolume = true
		} else {
			return nil, status.Errorf(codes.NotFound, "Error, volume %s not found", volName)
		}
	}

	err = ns.Mount.UnmountPath(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Unknown,
			"Error unpublishing volume on node, TargetPath: %s, error: %s", targetPath, err)
	}

	if ephemeralVolume {
		return nodeUnpublishEphemeral(ns, vol)
	}

	log.Infof("Successfully unpublish volume on node, targetPath: %s", targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeGetInfo(_ context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Infof("NodeGetInfo called with request %v", protosanitizer.StripSecrets(*req))

	nodeID, err := ns.Metadata.GetInstanceID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve instance id of node %s", err)
	}

	zone, err := ns.Metadata.GetAvailabilityZone()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve availability zone of node %v", err)
	}
	topology := &csi.Topology{Segments: map[string]string{topologyKey: zone}}

	return &csi.NodeGetInfoResponse{
		NodeId:             nodeID,
		AccessibleTopology: topology,
		MaxVolumesPerNode:  maxVolumes,
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
	if stats.Block {
		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{
				{
					Total: stats.TotalBytes,
					Unit:  csi.VolumeUsage_BYTES,
				},
			},
		}, nil
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{Total: stats.TotalBytes, Available: stats.AvailableBytes, Used: stats.UsedBytes, Unit: csi.VolumeUsage_BYTES},
			{Total: stats.TotalInodes, Available: stats.AvailableInodes, Used: stats.UsedInodes, Unit: csi.VolumeUsage_INODES},
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

func (ns *nodeServer) NodeExpandVolume(_ context.Context, req *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {
	log.Infof("NodeExpandVolume: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	volumePath := req.GetVolumePath()

	if err := nodeExpendValidation(ns.Driver.cloudCredentials, volumeID, volumePath); err != nil {
		return nil, err
	}

	output, err := ns.Mount.GetMountFs(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to find mount file system %s: %v", volumePath, err)
	}

	devicePath := strings.TrimSpace(string(output))
	if devicePath == "" {
		return nil, status.Errorf(codes.Unknown, "Unable to find device path for volume: %s", output)
	}

	r := mountutils.NewResizeFs(ns.Mount.Mounter().Exec)
	if _, err = r.Resize(devicePath, volumePath); err != nil {
		return nil, status.Errorf(codes.Unknown, "Could not resize volume %q:  %v", volumeID, err)
	}
	return &csi.NodeExpandVolumeResponse{}, nil
}

func nodeExpendValidation(cc *config.CloudCredentials, volumeID, volumePath string) error {
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, VolumeID not provided")
	}
	if len(volumePath) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, VolumePath not provided")
	}

	_, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return err
	}

	return nil
}

func collectMountOptions(fsType string, mntFlags []string) []string {
	var options []string
	options = append(options, mntFlags...)

	// By default, xfs does not allow mounting of two volumes with the same filesystem uuid.
	// Force ignore this uuid to be able to mount volume + its clone / restored snapshot on the same node.
	if fsType == "xfs" {
		options = append(options, "nouuid")
	}
	return options
}

func nodeUnpublishEphemeral(ns *nodeServer, vol *cloudvolumes.Volume) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Infof("nodeUnpublishEphemeral: called with args %v", protosanitizer.StripSecrets(*vol))

	cc := ns.Driver.cloudCredentials
	volumeID := vol.ID
	instanceID := ""

	if len(vol.Attachments) > 0 {
		instanceID = vol.Attachments[0].ServerID
	} else {
		return nil, status.Error(codes.FailedPrecondition, "Error, volume attachment not found in request")
	}

	err := services.DetachVolumeCompleted(cc, instanceID, volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error unpublishing ephemeral volume on node: %s", err)
	}

	err = services.DeleteVolume(cc, volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error deleting ephemeral volume: %s", err)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func nodePublishEphemeral(req *csi.NodePublishVolumeRequest, ns *nodeServer) (*csi.NodePublishVolumeResponse, error) {
	log.Infof("nodePublishEphemeral: called with args %v", protosanitizer.StripSecrets(*req))

	cc := ns.Driver.cloudCredentials
	size := 10 // default size is 1GB
	var err error

	volumeId := req.GetVolumeId()
	volumeName := fmt.Sprintf("ephemeral-%s", volumeId)
	volumeCapability := req.GetVolumeCapability()
	volAvailability, err := ns.Metadata.GetAvailabilityZone()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "retrieving availability zone from MetaData service failed with error %v", err)
	}

	capacity, ok := req.GetVolumeContext()["capacity"]
	if ok && strings.HasSuffix(capacity, "Gi") {
		size, err = strconv.Atoi(strings.TrimSuffix(capacity, "Gi"))
		if err != nil {
			return nil, status.Error(codes.Unknown, fmt.Sprintf("Unable to parse capacity %s: %v", capacity, err))
		}
	}

	// Check type in given param, if not, use ""
	volumeType := req.GetVolumeContext()["type"]

	metadata := map[string]string{"evs.csi.huaweicloud.com/cluster": ns.Driver.cluster}
	metadata[CreateForVolumeIDKey] = "true"
	metadata[DssIDKey] = req.VolumeContext[DssIDKey]

	volumeID, err := services.CreateVolumeCompleted(cc, &cloudvolumes.CreateOpts{
		Volume: cloudvolumes.VolumeOpts{
			Name:             volumeName,
			Size:             size,
			VolumeType:       volumeType,
			AvailabilityZone: volAvailability,
			Metadata:         metadata,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create ephemeral volume %v", err)
	}

	// attach volume
	instanceID, err := ns.Metadata.GetInstanceID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get instance ID with error %s", err)
	}

	err = services.AttachVolumeCompleted(cc, instanceID, volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to attach volume %s to ECS %s with error : %v",
			volumeID, instanceID, err)
	}

	m := ns.Mount
	devicePath, err := getDevicePath(ns.Driver.cloudCredentials, volumeID, m)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Unable to find devicePath for volume: %v", err))
	}

	targetPath := req.GetTargetPath()
	// Verify whether mounted
	notMnt, err := m.IsLikelyNotMountPointAttach(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Volume Mount
	if notMnt {
		// set default fstype is ext4
		log.Infof("Ephemeral not mount, set default fsType is ext4")
		fsType := "ext4"
		var options []string
		if mnt := volumeCapability.GetMount(); mnt != nil {
			if mnt.FsType != "" {
				fsType = mnt.FsType
			}
			mountFlags := mnt.GetMountFlags()
			options = append(options, collectMountOptions(fsType, mountFlags)...)
		}
		// Mount
		err = m.Mounter().FormatAndMount(devicePath, targetPath, fsType, options)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &csi.NodePublishVolumeResponse{}, nil

}

func nodePublishVolumeForBlock(req *csi.NodePublishVolumeRequest, ns *nodeServer, mountOptions []string) (
	*csi.NodePublishVolumeResponse, error) {
	log.Infof("nodePublishVolumeForBlock: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	podVolumePath := filepath.Dir(targetPath)

	m := ns.Mount

	// Do not trust the path provided by cinder, get the real path on node
	source, err := getDevicePath(ns.Driver.cloudCredentials, volumeID, m)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error query devicePath for volume %s: %v", volumeID, err)
	}

	exists, err := utilpath.Exists(utilpath.CheckFollowSymlink, podVolumePath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !exists {
		if err := m.MakeDir(podVolumePath); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not create dir %q, error: %v", podVolumePath, err)
		}
	}
	err = m.MakeFile(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error in making file %v", err)
	}

	if err := m.Mounter().Mount(source, targetPath, "", mountOptions); err != nil {
		if removeErr := os.Remove(targetPath); removeErr != nil {
			return nil, status.Errorf(codes.Internal, "Could not remove mount target %q: %v", targetPath, err)
		}
		return nil, status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, targetPath, err)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}
