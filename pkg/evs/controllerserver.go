package evs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chnsz/golangsdk/openstack/evs/v2/cloudvolumes"
	"github.com/chnsz/golangsdk/openstack/evs/v2/snapshots"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs/services"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
)

const (
	defaultSizeGB = 10
)

type ControllerServer struct {
	Driver *EvsDriver
}

func (cs *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse,
	error) {
	log.Infof("CreateVolumeCompleted: called with args %v", protosanitizer.StripSecrets(*req))
	credentials := cs.Driver.cloudCredentials

	volName := req.GetName()
	if err := createVolumeValidation(volName, req.GetVolumeCapabilities()); err != nil {
		return nil, err
	}

	// Volume Size - Default is 10 GiB
	sizeGB := defaultSizeGB
	if req.GetCapacityRange() != nil {
		sizeGB = int(utils.RoundUpSize(req.GetCapacityRange().GetRequiredBytes(), common.GbByteSize))
	}

	parameters := req.GetParameters()
	volumeAz := parameters["availability"]
	if len(volumeAz) == 0 {
		// Check from Topology
		if req.GetAccessibilityRequirements() != nil {
			volumeAz = getAZFromTopology(req.GetAccessibilityRequirements())
			log.Infof("Get AZ By GetAccessibilityRequirements availability zone: %s", volumeAz)
		}
	}

	dssID := parameters["dssId"]

	// Check if there are any volumes with the same name
	if vol, err := services.CheckVolumeExists(credentials, volName, sizeGB); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else if vol != nil {
		return buildCreateVolumeResponse(vol, dssID), nil
	}

	// Check if snapshot exists
	snapshotID, err := checkSnapshotExists(credentials, req.GetVolumeContentSource())
	if err != nil {
		return nil, err
	}

	iops := 0
	throughput := 0
	volumeType := parameters["type"]
	if volumeType == "GPSSD2" || volumeType == "ESSD2" {
		iops, throughput, err = getIoAndThrough(parameters)
		if err != nil {
			return nil, err
		}
	}

	metadata := cs.parseMetadata(req, snapshotID)
	createOpts := &cloudvolumes.CreateOpts{
		Volume: cloudvolumes.VolumeOpts{
			Name:             volName,
			Size:             sizeGB,
			VolumeType:       volumeType,
			AvailabilityZone: volumeAz,
			SnapshotID:       snapshotID,
			Metadata:         metadata,
			IOPS:             iops,
			Throughput:       throughput,
		},
		Scheduler: &cloudvolumes.SchedulerOpts{
			StorageID: dssID,
		},
	}

	volumeID := ""
	if sizeGB < 10 {
		volumeID, err = services.CreateCinderCompleted(credentials, createOpts)
	} else {
		volumeID, err = services.CreateVolumeCompleted(credentials, createOpts)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	volume, err := services.GetVolume(credentials, volumeID)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully created volume %s in Availability Zone: %s of size %d GiB",
		volume.ID, volume.AvailabilityZone, volume.Size)

	return buildCreateVolumeResponse(volume, dssID), nil
}

func getIoAndThrough(parameters map[string]string) (int, int, error) {
	var err error
	iops := 0
	throughput := 0

	iopsStr := parameters["iops"]
	if len(iopsStr) > 0 {
		iops, err = strconv.Atoi(iopsStr)
		if err != nil {
			return 0, 0, status.Errorf(codes.InvalidArgument, "iops error, expected a number, but got %s, error: %s", iopsStr, err)
		}
	}
	throughputStr := parameters["throughput"]
	if len(throughputStr) > 0 {
		throughput, err = strconv.Atoi(throughputStr)
		if err != nil {
			return 0, 0, status.Errorf(codes.InvalidArgument, "throughput error, expected a number, but got %s, error: %s", throughputStr, err)
		}
	}
	return iops, throughput, nil
}

func (cs *ControllerServer) parseMetadata(req *csi.CreateVolumeRequest, snapshotID string) map[string]string {
	parameters := req.GetParameters()
	scsi := parameters["scsi"]

	// build the metadata of create option
	metadata := make(map[string]string)
	metadata[CsiClusterNodeIDKey] = cs.Driver.nodeID
	metadata[CreateForVolumeIDKey] = "true"

	if snapshotID == "" && scsi != "" && (scsi == "true" || scsi == "false") {
		metadata[HwPassthroughKey] = scsi
	}

	if kmsID := parameters["kmsId"]; snapshotID == "" && kmsID != "" {
		metadata[CmkIDKey] = kmsID
		metadata[EncryptedKey] = "1"
	}

	for _, key := range []string{PvcNameTag, PvcNsTag, PvNameKey} {
		if v, ok := parameters[key]; ok {
			metadata[key] = v
		}
	}
	return metadata
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

func checkSnapshotExists(credentials *config.CloudCredentials, content *csi.VolumeContentSource) (string, error) {
	snapshotID := ""
	if content != nil && content.GetSnapshot() != nil {
		snapshotID = content.GetSnapshot().GetSnapshotId()
		_, err := services.GetSnapshot(credentials, snapshotID)
		if err != nil {
			if common.IsNotFound(err) {
				return snapshotID,
					status.Errorf(codes.NotFound, "Error, snapshot ID %s does not exist.", snapshotID)
			}
			return snapshotID,
				status.Errorf(codes.Internal, "Failed to retrieve the snapshot %s: %v", snapshotID, err)
		}
	}
	return snapshotID, nil
}

func (cs *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse,
	error) {
	log.Infof("DeleteVolume: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}

	credentials := cs.Driver.cloudCredentials
	if err := services.DeleteVolume(credentials, volumeID); err != nil {
		if common.IsNotFound(err) {
			log.Infof("Volume %s does not exist, skip deleting", volumeID)
			return &csi.DeleteVolumeResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "Error deleting volume: %s", err)
	}
	log.Infof("Successfully deleted volume %s", volumeID)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *ControllerServer) ControllerGetVolume(_ context.Context, req *csi.ControllerGetVolumeRequest) (
	*csi.ControllerGetVolumeResponse, error) {
	log.Infof("ControllerGetVolume: called with args %v", protosanitizer.StripSecrets(*req))

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}

	credentials := cs.Driver.cloudCredentials
	volume, err := services.GetVolume(credentials, volumeID)
	if err != nil {
		return nil, err
	}

	volStatus := &csi.ControllerGetVolumeResponse_VolumeStatus{}
	for _, attachment := range volume.Attachments {
		volStatus.PublishedNodeIds = append(volStatus.PublishedNodeIds, attachment.ServerID)
	}

	response := csi.ControllerGetVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: int64(volume.Size * common.GbByteSize),
		},
		Status: volStatus,
	}

	log.Infof("Successfully obtained volume details, volume ID: %s", volumeID)
	return &response, nil
}

type VolumeAttachmentStatus int32

const (
	VolumeNotAttached VolumeAttachmentStatus = iota + 1
	VolumeAttachingCurrentServer
	VolumeAttachingOtherServer
	VolumeAttachedCurrentServer
	VolumeAttachedOtherServer
	VolumeAttachError
)

func (cs *ControllerServer) ControllerPublishVolume(_ context.Context, req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {
	log.Infof("ControllerPublishVolume: called with args %+v", protosanitizer.StripSecrets(*req))
	credentials := cs.Driver.cloudCredentials
	instanceID := req.GetNodeId()
	volumeID := req.GetVolumeId()
	if err := publishValidation(credentials, volumeID, instanceID, req.GetVolumeCapability()); err != nil {
		return nil, err
	}
	volume, err := services.GetVolume(credentials, volumeID)
	if err != nil {
		return nil, err
	}

	attachmentStatus := volumeAttachmentStatus(volume, instanceID)
	log.Infof("ControllerPublishVolume: attachmentStatus is %s", attachmentStatus)
	switch attachmentStatus {
	case VolumeNotAttached:
		if err := services.AttachVolumeCompleted(credentials, instanceID, volumeID); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to publish volume %s to ECS %s with error %v",
				volumeID, instanceID, err)
		}
	case VolumeAttachingCurrentServer:
		if err := services.WaitForVolumeAttaching(credentials, volumeID); err != nil {
			return nil, status.Errorf(codes.Internal,
				"Failed to wait for volume: %s attaching ECS: %s with error %v", volumeID, instanceID, err)
		}
	case VolumeAttachingOtherServer:
		return nil, status.Errorf(codes.Internal, "Error, volume: %s is attaching another server", volumeID)
	case VolumeAttachedCurrentServer:
		log.Infof("ControllerPublishVolume volume: %s already attached on server: %s", volumeID, instanceID)
		return buildPublishVolumeResponse(volume, instanceID), nil
	case VolumeAttachedOtherServer:
		return nil, status.Errorf(codes.Internal, "Error, volume: %s is in used by another server", volumeID)
	default:
		return nil, status.Errorf(codes.Internal, "Error, status: %s was found in volume: %s, ",
			volume.Status, volume.ID)
	}

	log.Infof("Successfully published volume %s to EVS %s, obtaining device path", volumeID, instanceID)
	if volume, err = services.GetVolume(credentials, volumeID); err != nil {
		return nil, err
	}
	return buildPublishVolumeResponse(volume, instanceID), nil
}

func buildPublishVolumeResponse(volume *cloudvolumes.Volume, instanceID string) *csi.ControllerPublishVolumeResponse {
	devicePath := ""
	for _, attach := range volume.Attachments {
		if attach.ServerID == instanceID {
			devicePath = attach.Device
			break
		}
	}
	log.Infof("Got device path: %s", devicePath)
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			"DevicePath": devicePath,
		},
	}
}

func volumeAttachmentStatus(volume *cloudvolumes.Volume, instanceID string) VolumeAttachmentStatus {
	attachment := false // use to decide whether volume has attached on current instance
	for _, v := range volume.Attachments {
		if v.ServerID == instanceID {
			attachment = true
			break
		}
	}

	volumeStatus := volume.Status
	if services.EvsAvailableStatus == volumeStatus {
		return VolumeNotAttached
	}
	if services.EvsAttachingStatus == volumeStatus && attachment {
		return VolumeAttachingCurrentServer
	}
	if services.EvsAttachingStatus == volumeStatus && !attachment {
		return VolumeAttachingOtherServer
	}
	if services.EvsInUseStatus == volumeStatus && attachment {
		return VolumeAttachedCurrentServer
	}
	if services.EvsInUseStatus == volumeStatus && !attachment {
		return VolumeAttachedOtherServer
	}
	return VolumeAttachError
}

func publishValidation(cc *config.CloudCredentials, volumeID, instanceID string, capability *csi.VolumeCapability) error {
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}
	if len(instanceID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, ECS instance ID cannot be empty")
	}
	if capability == nil {
		return status.Error(codes.InvalidArgument, "Validation failed, volume capability cannot be empty")
	}

	if _, err := services.GetServer(cc, instanceID); err != nil {
		return err
	}

	return nil
}

func (cs *ControllerServer) ControllerUnpublishVolume(_ context.Context, req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {
	log.Infof("ControllerUnpublishVolume: called with args %v", protosanitizer.StripSecrets(*req))

	credentials := cs.Driver.cloudCredentials
	instanceID := req.GetNodeId()
	volumeID := req.GetVolumeId()

	volume, err := unpublishValidation(credentials, volumeID, instanceID)
	if err != nil {
		return nil, err
	}

	if volume.Status == services.EvsAvailableStatus || len(volume.Attachments) == 0 {
		log.Warningf("Warning, the volume %s is not in the server %s attach volume list, skip unpublishing",
			volumeID, instanceID)
		return &csi.ControllerUnpublishVolumeResponse{}, nil
	}

	err = services.DetachVolumeCompleted(credentials, instanceID, volumeID)
	if err != nil {
		if strings.Contains(err.Error(), "Ecs.0111") {
			log.Warningf("Warning, the volume %s is not in the server %s attach volume list, skip unpublishing",
				volumeID, instanceID)
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "Error unpublishing volume %s from server %s, error: %v",
			volumeID, instanceID, err)
	}

	log.Infof("Successfully unpublished, volume ID: %s", volumeID)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func unpublishValidation(cc *config.CloudCredentials, volumeID, instanceID string) (*cloudvolumes.Volume, error) {
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}
	if len(instanceID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, ECS instance ID cannot be empty")
	}

	volume, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return nil, err
	}

	if _, err = services.GetServer(cc, instanceID); err != nil {
		return nil, err
	}

	return volume, nil
}

func (cs *ControllerServer) ListVolumes(_ context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse,
	error) {
	log.Infof("ListVolumes: called with args %v", protosanitizer.StripSecrets(*req))

	if req.MaxEntries < 0 {
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("Validation failed, max entries request %v, must not be negative ", req.MaxEntries))
	}

	opts := cloudvolumes.ListOpts{
		Marker: req.StartingToken,
		Limit:  int(req.MaxEntries),
	}
	volumes, err := services.ListVolumes(cs.Driver.cloudCredentials, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	entries := make([]*csi.ListVolumesResponse_Entry, 0, len(volumes))
	for _, vol := range volumes {
		publishedNodeIds := make([]string, len(vol.Attachments))
		for _, attachment := range vol.Attachments {
			publishedNodeIds = append(publishedNodeIds, attachment.ServerID)
		}

		entry := csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				VolumeId:      vol.ID,
				CapacityBytes: int64(vol.Size * common.GbByteSize),
			},
			Status: &csi.ListVolumesResponse_VolumeStatus{
				PublishedNodeIds: publishedNodeIds,
			},
		}

		entries = append(entries, &entry)
	}

	response := &csi.ListVolumesResponse{
		Entries: entries,
	}

	log.Infof("Successfully obtained volume list, size: %v", len(entries))
	if len(volumes) > 0 && len(volumes) == opts.Limit {
		response.NextToken = volumes[len(volumes)-1].ID
	}

	return response, nil
}

func (cs *ControllerServer) CreateSnapshot(_ context.Context, req *csi.CreateSnapshotRequest) (
	*csi.CreateSnapshotResponse, error) {
	log.Infof("CreateSnapshot called with request %v", protosanitizer.StripSecrets(*req))

	credentials := cs.Driver.cloudCredentials
	name := req.GetName()
	volumeID := req.GetSourceVolumeId()
	if err := createSnapshotValidation(name, volumeID); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check create snapshot param, %v", err)
	}

	response, err := checkDuplicateSnapshotName(credentials, name, volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check duplicate snapshot name: %v", err)
	}
	if response != nil {
		log.Infof("Snapshot with name: %s | volumeID: %s already exist. detail: %v", name, volumeID, response)
		return response, nil
	}

	if _, err := services.GetVolume(credentials, volumeID); err != nil {
		return nil, err
	}
	snapshot, err := services.CreateSnapshotCompleted(credentials, name, volumeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create snapshot to completed: %v", err)
	}
	log.Infof("Successful create snapshot. detail: %v", protosanitizer.StripSecrets(snapshot))
	return buildSnapshotResponse(snapshot), nil
}

func buildSnapshotResponse(snap *snapshots.Snapshot) *csi.CreateSnapshotResponse {
	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     snap.ID,
			SizeBytes:      int64(snap.Size * common.GbByteSize),
			SourceVolumeId: snap.VolumeID,
			CreationTime:   timestamppb.New(snap.CreatedAt),
			ReadyToUse:     true,
		},
	}
}

func createSnapshotValidation(name string, volumeID string) error {
	if volumeID == "" {
		return status.Error(codes.InvalidArgument, "CreateSnapshot volumeID cannot be empty")
	}
	if name == "" {
		return status.Error(codes.InvalidArgument, "CreateSnapshot name cannot be empty")
	}
	return nil
}

func checkDuplicateSnapshotName(credentials *config.CloudCredentials, name string, volumeID string) (
	*csi.CreateSnapshotResponse, error) {
	listOpts := snapshots.ListOpts{
		Name: name,
	}
	pageList, err := services.ListSnapshots(credentials, listOpts)
	if err != nil {
		return nil, err
	}
	listSnapshots := pageList.Snapshots
	if len(listSnapshots) == 1 {
		snap := &listSnapshots[0]
		if snap.VolumeID != volumeID {
			return nil, status.Error(codes.AlreadyExists, "CreateSnapshot same name with different volumeId")
		}
		return buildSnapshotResponse(snap), nil
	}
	if len(listSnapshots) > 1 {
		return nil, status.Error(codes.Internal, "Multiple snapshots reported by EVS with same name")
	}
	return nil, nil
}

func (cs *ControllerServer) DeleteSnapshot(_ context.Context, req *csi.DeleteSnapshotRequest) (
	*csi.DeleteSnapshotResponse, error) {
	log.Infof("DeleteSnapshot called with request %v", protosanitizer.StripSecrets(*req))
	credentials := cs.Driver.cloudCredentials
	id := req.GetSnapshotId()
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "Snapshot ID must be provided in DeleteSnapshot request")
	}

	if err := services.DeleteSnapshot(credentials, id); err != nil {
		if common.IsNotFound(err) {
			log.Infof("Snapshot %s is already deleted.", id)
			return &csi.DeleteSnapshotResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "Failed to Delete snapshot: %v", err)
	}
	log.Infof("Successful delete snapshot")
	return &csi.DeleteSnapshotResponse{}, nil
}

func (cs *ControllerServer) ListSnapshots(_ context.Context, req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {
	log.Infof("ListSnapshots called with request %v", protosanitizer.StripSecrets(*req))
	credentials := cs.Driver.cloudCredentials

	availableStatus := "available"
	opts := snapshots.ListOpts{}
	if req.GetSnapshotId() != "" {
		opts.ID = req.GetSnapshotId()
	} else {
		opts.VolumeID = req.GetSourceVolumeId()
		opts.Status = availableStatus // Only retrieve snapshots available
		opts.Limit = int(req.MaxEntries)
		offset, err := strconv.Atoi(req.GetStartingToken())
		if err != nil {
			offset = 0
		}
		opts.Offset = offset
	}
	pageList, err := services.ListSnapshots(credentials, opts)
	if err != nil {
		return nil, err
	}

	var responses []*csi.ListSnapshotsResponse_Entry
	for _, element := range pageList.Snapshots {
		responses = append(responses, generateListSnapshotsResponseEntry(element))
	}
	response := &csi.ListSnapshotsResponse{Entries: responses}
	currentOffset := opts.Offset + len(responses)
	if currentOffset < pageList.Count {
		response.NextToken = strconv.Itoa(currentOffset)
	}
	log.Infof("Successful query snapshot list. detail: %v", protosanitizer.StripSecrets(response))
	return response, nil
}

func generateListSnapshotsResponseEntry(snapshot snapshots.Snapshot) *csi.ListSnapshotsResponse_Entry {
	snapshotEntry := csi.Snapshot{
		SizeBytes:      int64(snapshot.Size * common.GbByteSize),
		SnapshotId:     snapshot.ID,
		SourceVolumeId: snapshot.VolumeID,
		CreationTime:   timestamppb.New(snapshot.CreatedAt),
		ReadyToUse:     true,
	}
	return &csi.ListSnapshotsResponse_Entry{
		Snapshot: &snapshotEntry,
	}
}

// ControllerGetCapabilities implements the default GRPC callout.
// Default supports all capabilities
func (cs *ControllerServer) ControllerGetCapabilities(_ context.Context, _ *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.Driver.cscap,
	}, nil
}

func (cs *ControllerServer) ValidateVolumeCapabilities(_ context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {
	log.Infof("ValidateVolumeCapabilities: called with args %v", protosanitizer.StripSecrets(*req))

	volCapabilities := req.GetVolumeCapabilities()
	volumeID := req.GetVolumeId()

	if len(volCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume capabilities cannot be empty")
	}
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}
	if _, err := services.GetVolume(cs.Driver.cloudCredentials, volumeID); err != nil {
		return nil, err
	}

	for _, capability := range volCapabilities {
		if capability.GetAccessMode().GetMode() != cs.Driver.vcap[0].Mode {
			return &csi.ValidateVolumeCapabilitiesResponse{Message: "Requested volume capability not supported"}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: cs.Driver.vcap[0],
				},
			},
		},
	}, nil
}

func (cs *ControllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (
	*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCapacity is not yet implemented")
}

func (cs *ControllerServer) ControllerExpandVolume(_ context.Context, req *csi.ControllerExpandVolumeRequest) (
	*csi.ControllerExpandVolumeResponse, error) {
	log.Infof("ControllerExpandVolume: called with args %v", protosanitizer.StripSecrets(*req))
	cc := cs.Driver.cloudCredentials

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}
	capRange := req.GetCapacityRange()
	if capRange == nil {
		return nil, status.Error(codes.InvalidArgument, "Validation failed, capacity range cannot be empty")
	}

	sizeBytes := req.GetCapacityRange().GetRequiredBytes()
	sizeGB := int(utils.RoundUpSize(sizeBytes, common.GbByteSize))
	maxSizeBytes := capRange.GetLimitBytes()
	if maxSizeBytes > 0 && maxSizeBytes < sizeBytes {
		return nil, status.Errorf(codes.OutOfRange,
			"Validation failed, after round-up volume size %v exceeds the max size %v", sizeBytes, maxSizeBytes)
	}

	volume, err := services.GetVolume(cc, volumeID)
	if err != nil {
		return nil, err
	}
	if volume.Size >= sizeGB {
		log.Warningf("Volume %v has been already expanded to %v, requested %v", volumeID, volume.Size, sizeGB)
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         int64(volume.Size * common.GbByteSize),
			NodeExpansionRequired: true,
		}, nil
	}

	err = services.ExpandVolume(cc, volumeID, sizeGB)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Error resizing volume %v to size %v, error: %s", volumeID, sizeGB, err)
	}

	log.Infof("Successfully resized volume %v to size %v", volumeID, sizeGB)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         sizeBytes,
		NodeExpansionRequired: true,
	}, nil
}

func getAZFromTopology(requirement *csi.TopologyRequirement) string {
	for _, topology := range requirement.GetPreferred() {
		zone, exists := topology.GetSegments()[topologyKey]
		if exists {
			return zone
		}
	}

	for _, topology := range requirement.GetRequisite() {
		zone, exists := topology.GetSegments()[topologyKey]
		if exists {
			return zone
		}
	}
	return ""
}

func buildCreateVolumeResponse(vol *cloudvolumes.Volume, dssID string) *csi.CreateVolumeResponse {
	accessibleTopology := []*csi.Topology{
		{
			Segments: map[string]string{topologyKey: vol.AvailabilityZone},
		},
	}

	VolumeContext := make(map[string]string)
	if dssID != "" {
		VolumeContext[DssIDKey] = dssID
	}

	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:           vol.ID,
			CapacityBytes:      int64(vol.Size * common.GbByteSize),
			AccessibleTopology: accessibleTopology,
			VolumeContext:      VolumeContext,
		},
	}

	if vol.SnapshotID != "" {
		response.Volume.ContentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: vol.SnapshotID,
				},
			},
		}
	}

	if vol.SourceVolID != "" {
		response.Volume.ContentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Volume{
				Volume: &csi.VolumeContentSource_VolumeSource{
					VolumeId: vol.SourceVolID,
				},
			},
		}
	}
	return response
}
