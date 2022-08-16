package evs

import (
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	log "k8s.io/klog"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
	acceptance "github.com/huaweicloud/huaweicloud-csi-driver/test"

	"github.com/stretchr/testify/assert"
)

var (
	name    = fmt.Sprintf("k8s-csi-%s", utils.RandomString(5))
	size    = int64(10 * common.GbByteSize)
	newSize = int64(12 * common.GbByteSize)

	volumeType = "SSD"
	dssID      = ""
	scsi       = "false"

	cc     *config.CloudCredentials
	cs     *evs.ControllerServer
	driver *evs.EvsDriver

	ctx = context.Background()

	nodeID = acceptance.NodeID
)

func TestVolume(t *testing.T) {
	cc, err := acceptance.LoadConfig()
	if err != nil {
		log.Errorf("Error loading and verifying config data: %s", err)
	}
	driver = evs.NewDriver(cc, "unix://csi/csi.sock", "kubernetes", nodeID)
	cs = driver.GetControllerServer()

	volumeID := createVolume(t)
	defer deleteVolume(t, volumeID)
	readVolume(t, volumeID)
}

func createVolume(t *testing.T) string {
	req := &csi.CreateVolumeRequest{
		Name: name,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: size,
		},
		VolumeCapabilities: []*csi.VolumeCapability{
			{},
		},
		Parameters: map[string]string{
			"type":         volumeType,
			"availability": acceptance.Availability,
			"dssID":        dssID,
			"scsi":         scsi,
		},
	}

	createRsp, err := cs.CreateVolume(ctx, req)
	assert.Nilf(t, err, "Error creating volume, error: %s", err)
	log.Infof("Volume detail: %#v", createRsp)

	volumeID := createRsp.GetVolume().VolumeId
	assert.NotEmptyf(t, volumeID, "Error creating volume, we could not get the volume ID.")

	return volumeID
}

func readVolume(t *testing.T, volumeID string) {
	volume, err := cs.ControllerGetVolume(ctx, &csi.ControllerGetVolumeRequest{
		VolumeId: volumeID,
	})
	if err != nil && !common.IsNotFound(err) {
		t.Fatal("the evs volume is not exists")
	}
	assert.Equalf(t, volumeID, volume.Volume.VolumeId, "The EVS volume ID is not expected.")
	assert.Equalf(t, size, volume.Volume.CapacityBytes, "The EVS volume capacity is not expected.")
}

func deleteVolume(t *testing.T, volumeID string) {
	delReq := &csi.DeleteVolumeRequest{
		VolumeId: volumeID,
	}
	_, err := cs.DeleteVolume(context.Background(), delReq)
	assert.Nilf(t, err, "Error deleting volume, error: %s", err)

	err = common.WaitForCompleted(func() (bool, error) {
		rsp, err := cs.ControllerGetVolume(ctx, &csi.ControllerGetVolumeRequest{
			VolumeId: volumeID,
		})
		if rsp == nil && err != nil {
			return true, nil
		}
		return false, nil
	})
	assert.Nilf(t, err, "the evs volume is still exists")
}
