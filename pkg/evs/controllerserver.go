package evs

import (
	"fmt"

	"github.com/chnsz/golangsdk/openstack/evs/v2/cloudvolumes"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	log "k8s.io/klog"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs/services"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
)

type ControllerServer struct {
	Driver *EvsDriver
}

func (cs *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse,
	error) {
	log.Infof("CreateVolume called with args %+v", protosanitizer.StripSecrets(*req))

	volumeName := req.GetName()
	err := createValidation(volumeName, req.GetVolumeCapabilities())
	if err != nil {
		return nil, err
	}

	// Define the parameters for creation
	volSizeBytes := int64(1 * common.GbByteSize)
	if req.GetCapacityRange() != nil {
		volSizeBytes = req.GetCapacityRange().GetRequiredBytes()
	}
	volSizeGB := int(utils.RoundUpSize(volSizeBytes, common.GbByteSize))

	parameters := req.GetParameters()
	volType := parameters["type"]
	dssID := parameters["dssId"]
	scsi := parameters["scsi"]

	shareable := false
	if parameters["shareable"] == "true" {
		shareable = true
	}

	volAvailability := parameters["availability"]
	if len(volAvailability) == 0 {
		// Check from Topology
		if req.GetAccessibilityRequirements() != nil {
			volAvailability = common.GetAZFromTopology(req.GetAccessibilityRequirements(), driverName)
			log.Infof("Get AZ By GetAccessibilityRequirements Availability Zone: %s", volAvailability)
		}
	}

	cc := cs.Driver.cloudCredentials
	// Verify that a volume with the provided name exists for this tenant
	volumes, err := services.ListVolumes(cc, cloudvolumes.ListOpts{
		Name: volumeName,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to query the volume list, "+
			"unable to verify whether the volume exists, error: %s", err))
	}

	if len(volumes) == 1 {
		if volSizeGB != volumes[0].Size {
			return nil, status.Error(codes.AlreadyExists, "Create failed, volume Already exists with same name, "+
				"but different capacity")
		}
		log.Errorf("Volume %s already exists in Availability Zone: %s of size %d GiB",
			volumes[0].ID, volumes[0].AvailabilityZone, volumes[0].Size)
		return buildCreateVolumeRsp(&volumes[0], dssID, req.GetAccessibilityRequirements()), nil
	} else if len(volumes) > 1 {
		return nil, status.Error(codes.AlreadyExists,
			"Create failed, found multiple existing volumes with same name")
	}

	// build the metadata of create option
	metadata := make(map[string]string)
	metadata[CsiClusterNodeIDKey] = cs.Driver.nodeID
	metadata[CreateForVolumeIDKey] = "true"
	metadata[DssIDKey] = dssID

	if scsi != "" && (scsi == "true" || scsi == "false") {
		metadata[HwPassthroughKey] = scsi
	}
	for _, mKey := range []string{PvcNameTag, PvcNsTag, PvNameKey} {
		if v, ok := parameters[mKey]; ok {
			metadata[mKey] = v
		}
	}

	snapshotID := ""
	content := req.GetVolumeContentSource()
	if content != nil && content.GetSnapshot() != nil {
		snapshotID = content.GetSnapshot().GetSnapshotId()
		_, err = services.GetSnapshot(cc, snapshotID)
		if err != nil {
			if common.IsNotFound(err) {
				return nil, status.Errorf(codes.NotFound, "The snapshot(id: %s) does not exist", snapshotID)
			}
			return nil, status.Errorf(codes.Internal, "Failed to retrieve the snapshot %s: %v", snapshotID, err)
		}
	}

	// Create volume
	createOpts := cloudvolumes.CreateOpts{
		Volume: cloudvolumes.VolumeOpts{
			Name:             volumeName,
			Size:             volSizeGB,
			VolumeType:       volType,
			AvailabilityZone: volAvailability,
			SnapshotID:       snapshotID,
			Metadata:         metadata,
			Multiattach:      shareable,
		},
	}
	log.Infof("Create EVS volume options: %#v", createOpts)
	volumeID, err := services.CreateVolumeToCompletion(cc, createOpts)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Create EVS volume failed with error %v", err))
	}

	volume, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to query volume detail by id %s: %v",
			volumeID, err))
	}

	log.Infof("CreateVolume: Successfully created volume %s in Availability Zone: %s of size %d GiB",
		volume.ID, volume.AvailabilityZone, volume.Size)
	return buildCreateVolumeRsp(volume, dssID, req.GetAccessibilityRequirements()), nil
}

func createValidation(volumeName string, volCapabilities []*csi.VolumeCapability) error {
	if len(volumeName) == 0 {
		log.Errorf("Volume capabilities cannot be empty")
		return status.Error(codes.InvalidArgument, "EVS volume name cannot be empty")
	}

	if volCapabilities == nil {
		log.Errorf("Volume capabilities cannot be empty")
		return status.Error(codes.InvalidArgument, "Volume capabilities cannot be empty")
	}

	return nil
}

func buildCreateVolumeRsp(vol *cloudvolumes.Volume, dssID string, accessibleTopologyReq *csi.TopologyRequirement) *csi.
	CreateVolumeResponse {
	var contentSource *csi.VolumeContentSource
	if vol.SnapshotID != "" {
		contentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: vol.SnapshotID,
				},
			},
		}
	}

	if vol.SourceVolID != "" {
		contentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Volume{
				Volume: &csi.VolumeContentSource_VolumeSource{
					VolumeId: vol.SourceVolID,
				},
			},
		}
	}

	accessibleTopology := []*csi.Topology{
		{
			Segments: map[string]string{topologyKey: vol.AvailabilityZone},
		},
	}

	VolumeContext := make(map[string]string)
	if dssID != "" {
		VolumeContext[DssIDKey] = dssID
	}
	resp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:           vol.ID,
			CapacityBytes:      int64(vol.Size * common.GbByteSize),
			AccessibleTopology: accessibleTopology,
			ContentSource:      contentSource,
			VolumeContext:      VolumeContext,
		},
	}
	return resp
}

func (cs *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse,
	error) {
	log.Infof("DeleteVolume: called with args %+v", protosanitizer.StripSecrets(*req))
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume: Volume ID cannot be empty")
	}

	if err := services.DeleteVolume(cs.Driver.cloudCredentials, volumeID); err != nil {
		if common.IsNotFound(err) {
			log.Infof("Volume %s does not exist, skip deleting", volumeID)
			return &csi.DeleteVolumeResponse{}, nil
		}
		return nil, status.Error(codes.Internal,
			fmt.Sprintf("Failed to delete volume, id: %s, error: %v", volumeID, err))
	}
	log.Infof("DeleteVolume: Successfully deleted volume %s", volumeID)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerGetVolume(_ context.Context, req *csi.ControllerGetVolumeRequest) (
	*csi.ControllerGetVolumeResponse, error) {
	klog.Infof("ControllerGetVolume: called with args %+v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ControllerGetVolume: Volume ID cannot be empty")
	}

	volume, err := services.GetVolume(cs.Driver.cloudCredentials, volumeID)
	if err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "Volume %s not found", volumeID)
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("ControllerGetVolume failed with error %v", err))
	}

	volumeStatus := &csi.ControllerGetVolumeResponse_VolumeStatus{}
	for _, attachment := range volume.Attachments {
		volumeStatus.PublishedNodeIds = append(volumeStatus.PublishedNodeIds, attachment.ServerID)
	}

	return &csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: int64(volume.Size * 1024 * 1024 * 1024),
		},
		Status: volumeStatus,
	}, nil
}

func (cs *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ControllerUnpublishVolume(_ context.Context, req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ListVolumes(_ context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse,
	error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) CreateSnapshot(_ context.Context, req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse,
	error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities implements the default GRPC callout.
// Default supports all capabilities
func (cs *ControllerServer) ControllerGetCapabilities(_ context.Context, req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *ControllerServer) ValidateVolumeCapabilities(_ context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *ControllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
