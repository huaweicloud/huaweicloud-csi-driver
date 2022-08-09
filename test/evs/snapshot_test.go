package evs

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/evs"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
	acceptance "github.com/huaweicloud/huaweicloud-csi-driver/test"
	"testing"
	"time"
)

var (
	snapshotName = fmt.Sprintf("k8s-test-snapshot-%s", utils.RandomString(5)) //云硬盘快照名称
	maxEntries   = 100                                                        // 云硬盘每页条目数量
)

func TestSnapshot(t *testing.T) {
	cc, err := acceptance.LoadConfig()
	if err != nil {
		t.Errorf("Error loading and verifying config data: %s", err)
	}
	driver = evs.NewDriver(cc, "unix://csi/csi.sock", "kubernetes", nodeID)
	cs = driver.GetControllerServer()

	volumeID := createVolume(t)
	defer deleteVolume(t, volumeID)

	snapshot := createSnapshot(t, volumeID)
	listSnapshots(t, snapshot.Snapshot.SnapshotId)
	defer deleteSnapshot(t, snapshot.Snapshot.SnapshotId)
}

func createSnapshot(t *testing.T, sourceVolumeId string) *csi.CreateSnapshotResponse {
	req := csi.CreateSnapshotRequest{
		SourceVolumeId: sourceVolumeId,
		Name:           snapshotName,
	}

	response, err := cs.CreateSnapshot(ctx, &req)
	if err != nil {
		t.Errorf("UT createSnapshot response Error, %v", err)
	}
	t.Logf("UT createSnapshot detail is %v", response)
	if response.Snapshot.SnapshotId == "" {
		t.Errorf("UT createSnapshot response snapshotId can not be Empty Error, %v", err)
	}

	listReq := csi.ListSnapshotsRequest{
		SnapshotId: response.Snapshot.SnapshotId,
	}
	listResponse, err := cs.ListSnapshots(ctx, &listReq)
	if err != nil {
		t.Errorf("UT createSnapshot query listSnapshot Error, %v", err)
	}
	if len(listResponse.Entries) != 1 || listResponse.Entries[0].Snapshot.SnapshotId != response.Snapshot.SnapshotId {
		t.Errorf("UT createSnapshot failed, for can not query snapshotId, %s", response.Snapshot.SnapshotId)
	}
	return response
}

func listSnapshots(t *testing.T, snapshotId string) {
	req := csi.ListSnapshotsRequest{
		SnapshotId: snapshotId,
	}
	response, err := cs.ListSnapshots(ctx, &req)
	if err != nil {
		t.Errorf("UT listSnapshots each error, %v", err)
	}
	if len(response.Entries) != 1 || response.Entries[0].Snapshot.SnapshotId != snapshotId {
		t.Errorf("UT listSnapshots failed, %v", response)
	}
	t.Logf("UT listSnapshots each detail, %v", response)
	req = csi.ListSnapshotsRequest{
		MaxEntries: int32(maxEntries),
	}
	responses, err := cs.ListSnapshots(ctx, &req)
	if err != nil {
		t.Errorf("UT listSnapshots page error, %v", err)
	}
	if len(responses.Entries) == 0 {
		t.Errorf("UT listSnapshots page size zero")
	}
	t.Logf("UT listSnapshots page details, %v", responses)

}

func deleteSnapshot(t *testing.T, snapshotId string) {
	req := csi.DeleteSnapshotRequest{
		SnapshotId: snapshotId,
	}
	response, err := cs.DeleteSnapshot(ctx, &req)
	if err != nil {
		t.Errorf("UT deleteSnapshot error, %v", err)
	}
	t.Logf("UT deleteSnapshot response detail, %v", response)
	// takes a while to take effect
	time.Sleep(10 * time.Second)

	listReq := csi.ListSnapshotsRequest{
		SnapshotId: snapshotId,
	}
	listResponse, err := cs.ListSnapshots(ctx, &listReq)
	if err != nil {
		t.Errorf("UT deleteSnapshot query listSnapshot Error, %v", err)
	}
	if len(listResponse.Entries) > 0 {
		t.Errorf("UT deleteSnapshot Failed, for the deleted snapshot was found")
	}

}
