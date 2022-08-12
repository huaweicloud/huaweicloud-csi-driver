package evs

import (
	"fmt"
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

type ControllerServer struct {
	Driver *EvsDriver
}

func (cs *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse,
	error) {
	log.Infof("CreateVolumeCompleted: called with args %+v", protosanitizer.StripSecrets(*req))
	credentials := cs.Driver.cloudCredentials

	volName := req.GetName()
	if err := createValidation(volName, req.GetVolumeCapabilities()); err != nil {
		return nil, err
	}

	// Volume Size - Default is 10 GiB
	sizeGB := 10
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
	shareable := false
	if parameters["shareable"] == "true" {
		shareable = true
	}
	volumeType := parameters["type"]
	dssID := parameters["dssId"]
	scsi := parameters["scsi"]

	// Check if there are any volumes with the same name
	if vol, err := services.CheckVolumeExists(credentials, volName, sizeGB); err != nil {
		return nil, status.Errorf(codes.Internal,
			"Error querying volume details by name, unable to verify existence: %s", err)
	} else if vol != nil {
		return getCreateVolumeResponse(vol, dssID), nil
	}

	// Check if snapshot exists
	snapshotID, err := checkSnapshotExists(credentials, req.GetVolumeContentSource())
	if err != nil {
		return nil, err
	}

	// build the metadata of create option
	metadata := make(map[string]string)
	metadata[CsiClusterNodeIDKey] = cs.Driver.nodeID
	metadata[CreateForVolumeIDKey] = "true"
	metadata[DssIDKey] = dssID

	if scsi != "" && (scsi == "true" || scsi == "false") {
		metadata[HwPassthroughKey] = scsi
	}

	for _, key := range []string{PvcNameTag, PvcNsTag, PvNameKey} {
		if v, ok := parameters[key]; ok {
			metadata[key] = v
		}
	}

	volumeID, err := services.CreateVolumeCompleted(credentials, &cloudvolumes.CreateOpts{
		Volume: cloudvolumes.VolumeOpts{
			Name:             volName,
			Size:             sizeGB,
			VolumeType:       volumeType,
			AvailabilityZone: volumeAz,
			SnapshotID:       snapshotID,
			Metadata:         metadata,
			Multiattach:      shareable,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	volume, err := services.GetVolume(credentials, volumeID)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully created volume %s in Availability Zone: %s of size %d GiB",
		volume.ID, volume.AvailabilityZone, volume.Size)

	return getCreateVolumeResponse(volume, dssID), nil
}

func createValidation(volumeName string, capabilities []*csi.VolumeCapability) error {
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
	log.Infof("DeleteVolume: called with args %+v", protosanitizer.StripSecrets(*req))

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
	log.Infof("ControllerGetVolume: called with args %+v", protosanitizer.StripSecrets(*req))

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

func (cs *ControllerServer) ControllerPublishVolume(_ context.Context, req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {
	log.Infof("ControllerPublishVolume: called with args %+v", protosanitizer.StripSecrets(*req))

	credentials := cs.Driver.cloudCredentials
	instanceID := req.GetNodeId()
	volumeID := req.GetVolumeId()
	if err := publishValidation(credentials, volumeID, instanceID, req.GetVolumeCapability()); err != nil {
		return nil, err
	}

	if err := services.AttachVolumeCompleted(credentials, instanceID, volumeID); err != nil {
		if strings.Contains(err.Error(), "Ecs.0005") {
			log.Warningf("Warning, duplicate publish volume, skip publish volume")
			return nil, nil
		}
		return nil, status.Errorf(codes.Internal, "Failed to publish volume %s to ECS %s with error %v",
			volumeID, instanceID, err)
	}

	log.Infof("Successfully published volume %s to EVS %s, obtaining device path", volumeID, instanceID)

	volume, err := services.GetVolume(credentials, volumeID)
	if err != nil {
		return nil, err
	}
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
	}, nil
}

func publishValidation(cc *config.CloudCredentials, volumeID, instanceID string,
	capability *csi.VolumeCapability) error {
	if len(volumeID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, volume ID cannot be empty")
	}
	if len(instanceID) == 0 {
		return status.Error(codes.InvalidArgument, "Validation failed, ECS instance ID cannot be empty")
	}
	if capability == nil {
		return status.Error(codes.InvalidArgument, "Validation failed, volume capability cannot be empty")
	}

	if _, err := services.GetVolume(cc, volumeID); err != nil {
		return err
	}

	if _, err := services.GetServer(cc, instanceID); err != nil {
		return err
	}

	return nil
}

func (cs *ControllerServer) ControllerUnpublishVolume(_ context.Context, req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {
	log.Infof("ControllerUnpublishVolume: called with args %+v", protosanitizer.StripSecrets(*req))

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
	log.Infof("ListVolumes: called with args %+v", protosanitizer.StripSecrets(*req))

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
	volumeId := req.GetSourceVolumeId()
	if err := checkCreateSnapshotParamEmpty(name, volumeId); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check create snapshot param, %v", err)
	}
	response, err := checkDuplicateSnapshotName(credentials, name, volumeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check duplicate snapshot name: %v", err)
	}
	if response != nil {
		log.Infof("Snapshot with name: %s | volumeId: %s already exist. detail: %v", name, volumeId, response)
		return response, nil
	}
	if err := checkVolumeIsExist(credentials, volumeId); err != nil {
		return nil, status.Errorf(codes.Internal,
			"Failed to check volume is exist. volumeId: %s, error: %v", volumeId, err)
	}
	snapshot, err := services.CreateSnapshotToCompleted(credentials, name, volumeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create snapshot to completed: %v", err)
	}
	log.Infof("Successful create snapshot. detail: %v", protosanitizer.StripSecrets(snapshot))
	return generateSnapshotResponse(snapshot), nil
}

func generateSnapshotResponse(snap *snapshots.Snapshot) *csi.CreateSnapshotResponse {
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

func checkCreateSnapshotParamEmpty(name string, volumeId string) error {
	if volumeId == "" {
		return status.Error(codes.InvalidArgument, "CreateSnapshot volumeId cannot be empty")
	}
	if name == "" {
		return status.Error(codes.InvalidArgument, "CreateSnapshot name cannot be empty")
	}
	return nil
}

func checkDuplicateSnapshotName(credentials *config.CloudCredentials, name string, volumeId string) (
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
		if snap.VolumeID != volumeId {
			return nil, status.Error(codes.AlreadyExists, "CreateSnapshot same name with different volumeId")
		}
		return generateSnapshotResponse(snap), nil
	}
	if len(listSnapshots) > 1 {
		return nil, status.Error(codes.Internal, "Multiple snapshots reported by EVS with same name")
	}
	return nil, nil
}

func checkVolumeIsExist(credentials *config.CloudCredentials, volumeId string) error {
	volume, err := services.GetVolume(credentials, volumeId)
	if err != nil {
		return err
	}
	if volume == nil {
		return status.Error(codes.NotFound, "Volume not exist")
	}
	return nil
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

	var response *csi.ListSnapshotsResponse
	var err error
	snapshotId := req.GetSnapshotId()
	if snapshotId != "" {
		response, err = querySnapshotBySnapshotId(credentials, snapshotId)
	} else {
		response, err = querySnapshotPageList(credentials, req)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to query snapshots list: %v", err)
	}
	log.Infof("Successful query snapshot list. detail: %v", protosanitizer.StripSecrets(response))
	return response, nil
}

func querySnapshotBySnapshotId(credentials *config.CloudCredentials, id string) (*csi.ListSnapshotsResponse, error) {
	snapshot, err := services.GetSnapshot(credentials, id)
	if err != nil {
		if common.IsNotFound(err) {
			log.Infof("Snapshot %s not found", id)
			return &csi.ListSnapshotsResponse{}, nil
		}
		return nil, err
	}
	responseEntry := generateListSnapshotsResponseEntry(snapshot)
	return &csi.ListSnapshotsResponse{Entries: []*csi.ListSnapshotsResponse_Entry{responseEntry}}, nil
}

func querySnapshotPageList(credentials *config.CloudCredentials, req *csi.ListSnapshotsRequest) (
	*csi.ListSnapshotsResponse, error) {
	availableStatus := "available"
	opts := snapshots.ListOpts{
		VolumeID: req.GetSourceVolumeId(),
		Status:   availableStatus, // Only retrieve snapshots available
		Limit:    int(req.MaxEntries),
	}
	pageList, err := services.ListSnapshots(credentials, opts)
	if err != nil {
		return nil, err
	}

	var responses []*csi.ListSnapshotsResponse_Entry
	for _, element := range pageList.Snapshots {
		responses = append(responses, generateListSnapshotsResponseEntry(&element))
	}
	return &csi.ListSnapshotsResponse{Entries: responses}, nil
}

func generateListSnapshotsResponseEntry(snapshot *snapshots.Snapshot) *csi.ListSnapshotsResponse_Entry {
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
	log.Infof("ValidateVolumeCapabilities: called with args %+v", protosanitizer.StripSecrets(*req))

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
	log.Infof("ControllerExpandVolume: called with args %+v", protosanitizer.StripSecrets(*req))
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

func getCreateVolumeResponse(vol *cloudvolumes.Volume, dssID string) *csi.CreateVolumeResponse {
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
