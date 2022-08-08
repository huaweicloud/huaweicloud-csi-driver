package evs

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs/services"
	acceptance "github.com/huaweicloud/huaweicloud-csi-driver/test"
	"testing"
)

func TestControllerVolume(t *testing.T) {
	cc, err := acceptance.LoadConfig()
	if err != nil {
		t.Errorf("Error loading and verifying config data: %s", err)
	}
	driver = evs.NewDriver(cc, "unix://csi/csi.sock", "kubernetes", nodeID)
	cs = driver.GetControllerServer()

	volumeID := createVolume(t)
	defer deleteVolume(t, volumeID)

	validateVolumeCapabilities(t, volumeID)
	controllerGetVolume(t, volumeID)
	controllerExpandVolume(t, volumeID)
	controllerPublishVolume(t, volumeID, "904826a6-4ea2-4ff3-9c60-d55c3161ffb0")
	controllerUnpublishVolume(t, volumeID, "904826a6-4ea2-4ff3-9c60-d55c3161ffb0")

}

func validateVolumeCapabilities(t *testing.T, volumeId string) {
	supportCapability := cs.Driver.GetVolumeCapabilityAccessModes()[0].GetMode()
	unSupportCapability := csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY
	// this api supports one mode only
	// check unsupported capability
	volumeCapability := csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: unSupportCapability,
		},
	}
	req := csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           volumeId,
		VolumeCapabilities: []*csi.VolumeCapability{&volumeCapability},
	}
	capabilities, err := cs.ValidateVolumeCapabilities(ctx, &req)
	if err != nil {
		t.Errorf("UT validateVolumeCapabilities Error, %v", err)
	}
	if confirmed := capabilities.Confirmed; confirmed == nil {
		t.Logf("UT validateVolumeCapabilities colume capability not supported")
	}

	// check supported capability
	volumeCapability = csi.VolumeCapability{
		AccessMode: &csi.VolumeCapability_AccessMode{
			Mode: supportCapability,
		},
	}
	req = csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           volumeId,
		VolumeCapabilities: []*csi.VolumeCapability{&volumeCapability},
	}
	capabilities, err = cs.ValidateVolumeCapabilities(ctx, &req)
	if err != nil {
		t.Errorf("UT validateVolumeCapabilities Error, %v", err)
	}
	volumeCapabilities := capabilities.Confirmed.VolumeCapabilities
	if len(volumeCapabilities) != 1 || volumeCapabilities[0].AccessMode != cs.Driver.GetVolumeCapabilityAccessModes()[0] {
		t.Errorf("UT validateVolumeCapabilities capability Error empty")
	}
	t.Logf("UT validateVolumeCapabilities detail : %v", capabilities)
}

func controllerGetVolume(t *testing.T, volumeId string) {
	volume, err := services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerGetVolume get volume Error, %v", err)
	}
	t.Logf("UT controllerGetVolume volume detail, %v", volume)

	volumeResponse, err := cs.ControllerGetVolume(ctx, &csi.ControllerGetVolumeRequest{
		VolumeId: volumeId,
	})
	if err != nil {
		t.Errorf("UT controllerGetVolume Error, %v", err)
	}
	t.Logf("UT controllerGetVolume volumeResponse detail, %v", volumeResponse)
	publishNodeIds := volumeResponse.Status.PublishedNodeIds
	m := make(map[string]string, len(publishNodeIds))
	for _, value := range publishNodeIds {
		m[value] = ""
	}
	for _, attachment := range volume.Attachments {
		serverId := attachment.ServerID
		if _, ok := m[serverId]; ok {
			continue
		}
		t.Errorf("UT controllerGetVolume did not match serverId")
		return
	}

}

// 系统盘在购买云服务器时自动购买并挂载，无法单独购买，系统盘最大的容量是1024GB
// 数据盘可以在购买云服务器时购买，由系统自动挂载给云服务器。也可以在购买了云服务器之后，单独购买云硬盘并挂载给云服务器。数据盘的最大容量为32768GB
// 当前UT，只能以数据盘容量进行测试
var (
	gbToByteMultiple = 1024 * 1024 * 1024 // GB转Byte的倍数
	maxSize          = 32768              // GB
)

func controllerExpandVolume(t *testing.T, volumeId string) {
	// test expand volume 10G
	volume, err := services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerExpandVolume get volume Error, %v", err)
	}
	t.Logf("UT controllerExpandVolume volume detail, %v", volume)

	// correct expand volume
	beforeSize := volume.Size       // GB
	requiredSize := beforeSize + 10 // GB
	req := csi.ControllerExpandVolumeRequest{
		VolumeId: volumeId,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(requiredSize * gbToByteMultiple),
			LimitBytes:    int64(maxSize * gbToByteMultiple),
		},
	}
	expandVolume, err := cs.ControllerExpandVolume(ctx, &req)
	if err != nil {
		t.Errorf("UT controllerExpandVolume first Error, %v", expandVolume)
	}
	t.Logf("UT controllerExpandVolume first expand after detail, %v", expandVolume)
	volume, err = services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerExpandVolume first expand after get volume Error, %v", err)
	}
	t.Logf("UT controllerExpandVolume first expand after volume detail, %v", volume)
	if volume.Size*gbToByteMultiple != int(expandVolume.CapacityBytes) {
		t.Errorf("UT controllerExpandVolume first Failed, cause expand volume size not equal to requiredSize")
	}

	// expand size equals volume size
	beforeSize = volume.Size
	requiredSize = beforeSize
	req = csi.ControllerExpandVolumeRequest{
		VolumeId: volumeId,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(requiredSize * gbToByteMultiple),
			LimitBytes:    int64(maxSize * gbToByteMultiple),
		},
	}
	expandVolume, err = cs.ControllerExpandVolume(ctx, &req)
	if err != nil {
		t.Errorf("UT controllerExpandVolume second Error, %v", expandVolume)
	}
	t.Logf("UT controllerExpandVolume second expand after detail, %v", expandVolume)
	volume, err = services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerExpandVolume second expand after get volume Error, %v", err)
	}
	t.Logf("UT controllerExpandVolume expand after volume detail, %v", volume)
	if volume.Size*gbToByteMultiple != int(expandVolume.CapacityBytes) {
		t.Errorf("UT controllerExpandVolume second Failed, cause expand volume size not equal to requiredSize")
	}
}

func controllerPublishVolume(t *testing.T, volumeId string, nodeId string) {
	volume, err := services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerPublishVolume get volume Error, %v", err)
	}
	t.Logf("UT controllerPublishVolume volume detail, %v", volume)
	for _, v := range volume.Attachments {
		if v.AttachmentID == nodeId {
			// volumeId has already attached nodeId
			t.Errorf("UT controllerPublishVolume need to change param value")
			return
		}
	}

	accessMode := csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
	volumeCapability := csi.VolumeCapability{
		AccessType: nil,
		AccessMode: &accessMode,
	}
	req := csi.ControllerPublishVolumeRequest{
		NodeId:           nodeId,
		VolumeId:         volumeId,
		VolumeCapability: &volumeCapability,
	}
	volumeResponse, err := cs.ControllerPublishVolume(ctx, &req)
	if err != nil {
		t.Errorf("UT controllerPublishVolume Error, %v", err)
	}
	t.Logf("UT controllerPublishVolume detail, %v", volumeResponse)

	// check whether volume was successfully attached
	volume, err = services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerPublishVolume after attached get volume Error, %v", err)
	}
	t.Logf("UT controllerPublishVolume after attached volume detail, %v", volume)
	for _, v := range volume.Attachments {
		if v.ServerID == nodeId {
			// success attached
			return
		}
	}
	t.Errorf("UT controllerPublishVolume attached Failed, cause attachmentId did not matched")
}

func controllerUnpublishVolume(t *testing.T, volumeId string, nodeId string) {
	volume, err := services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerUnpublishVolume get volume Error, %v", err)
	}
	t.Logf("UT controllerUnpublishVolume volume detail, %v", volume)
	boo := false
	for _, v := range volume.Attachments {
		if v.ServerID == nodeId {
			boo = true
			break
		}
	}
	if !boo {
		t.Errorf("UT controllerUnpublishVolume need to change param value")
		return
	}

	req := csi.ControllerUnpublishVolumeRequest{
		VolumeId: volumeId,
		NodeId:   nodeId,
	}
	volumeResponse, err := cs.ControllerUnpublishVolume(ctx, &req)
	if err != nil {
		t.Errorf("UT controllerUnpublishVolume Error, %v", volumeResponse)
	}
	t.Errorf("UT controllerUnpublishVolume detail, %v", volumeResponse)

	// check whether volume was successfully unAttached
	volume, err = services.GetVolume(cc, volumeId)
	if err != nil {
		t.Errorf("UT controllerUnpublishVolume after unAttached get volume Error, %v", err)
	}
	t.Logf("UT controllerUnpublishVolume after unAttached volume detail, %v", volume)
	for _, v := range volume.Attachments {
		if v.AttachmentID == nodeId {
			// did not success unAttached
			t.Errorf("UT controllerUnpublishVolume unAttached Failed, cause attachmentId still bind nodeId")
			return
		}
	}
}
