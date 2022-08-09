package services

import (
	"github.com/chnsz/golangsdk/openstack/evs/v2/snapshots"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/klog/v2"
)

const debugPrefix = "[DEBUG] "

func GetSnapshot(c *config.CloudCredentials, id string) (*snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	return snapshots.Get(client, id).Extract()
}

func ListSnapshots(c *config.CloudCredentials, opts snapshots.ListOpts) ([]snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}
	page, err := snapshots.ListPage(client, opts)
	if err != nil {
		return nil, err
	}
	return page.Snapshots, nil
}

func CreateSnapshotToCompleted(credentials *config.CloudCredentials, name string, volumeId string) (
	*csi.CreateSnapshotResponse, error) {
	opts := &snapshots.CreateOpts{
		VolumeID: volumeId,
		Name:     name,
	}
	klog.V(4).Infof(debugPrefix+"createSnapshot opts: %v", *opts)
	snap, err := CreateSnapshot(credentials, opts)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof(debugPrefix+"createSnapshot response detail: %v", snap)
	err = WaitSnapshotReady(credentials, snap.ID)
	if err != nil {
		return nil, err
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     snap.ID,
			SizeBytes:      int64(snap.Size * 1024 * 1024 * 1024),
			SourceVolumeId: snap.VolumeID,
			CreationTime:   timestamppb.New(snap.CreatedAt),
			ReadyToUse:     true,
		},
	}, nil
}

func CreateSnapshot(c *config.CloudCredentials, opts *snapshots.CreateOpts) (*snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}
	return snapshots.Create(client, opts).Extract()
}

func WaitSnapshotReady(c *config.CloudCredentials, snapshotId string) error {
	availableStatus := "available"
	creatingStatus := "creating"
	err := common.WaitForCompleted(func() (done bool, err error) {
		snapshot, err := GetSnapshot(c, snapshotId)
		if err != nil {
			klog.Errorf("Failed to query snapshot when wait snapshot ready: %v", err)
			return false, err
		}
		klog.V(4).Infof(debugPrefix+"query snapshot detail when wait snapshot ready detail: %v", snapshot)
		if snapshot.Status == availableStatus {
			return true, nil
		}
		if snapshot.Status == creatingStatus {
			return false, nil
		}
		return false, status.Error(codes.Internal, "created snapshot status is not available")
	})
	return err
}

func DeleteSnapshot(c *config.CloudCredentials, snapshotId string) error {
	client, err := getEvsV2Client(c)
	if err != nil {
		return err
	}
	return snapshots.Delete(client, snapshotId).ExtractErr()
}

func List(c *config.CloudCredentials, opts snapshots.ListOpts) ([]snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	page, err := snapshots.ListPage(client, opts)
	if err != nil {
		klog.Errorf("Failed to query snapshot list page: %v", err)
		return nil, err
	}
	klog.V(4).Infof(debugPrefix+"query snapshot list page detail: %v", page)
	return page.Snapshots, nil
}
