package services

import (
	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/evs/v2/snapshots"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

// HCS does not publish /v2/{project_id}/cloudsnapshots, so with hcs set we call /v2/{project_id}/snapshots

func GetSnapshot(c *config.CloudCredentials, id string) (*snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	if !c.Global.Hcs {
		return snapshots.Get(client, id).Extract()
	}

	var r snapshots.GetResult
	_, r.Err = client.Get(client.ServiceURL("snapshots", id), &r.Body, nil)
	return r.Extract()
}

func ListSnapshots(c *config.CloudCredentials, opts snapshots.ListOpts) (*snapshots.PagedList, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	var page *snapshots.PagedList
	if c.Global.Hcs {
		page, err = listNativeSnapshots(c, client, opts)
	} else {
		page, err = snapshots.ListPage(client, opts)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to query snapshot list page: %v", err)
	}
	log.V(4).Infof("[DEBUG] query snapshot list page detail: %v", page)
	return page, nil
}

func listNativeSnapshots(c *config.CloudCredentials, client *golangsdk.ServiceClient, opts snapshots.ListOpts) (
	*snapshots.PagedList, error) {
	// no id filter on this list
	if opts.ID != "" {
		snap, err := GetSnapshot(c, opts.ID)
		if err != nil {
			if common.IsNotFound(err) {
				return &snapshots.PagedList{}, nil
			}
			return nil, err
		}
		return &snapshots.PagedList{Count: 1, Snapshots: []snapshots.Snapshot{*snap}}, nil
	}

	query, err := opts.ToSnapshotListQuery()
	if err != nil {
		return nil, err
	}
	var r golangsdk.Result
	_, r.Err = client.Get(client.ServiceURL("snapshots", "detail")+query, &r.Body, nil)
	var page snapshots.PagedList
	if err := r.ExtractInto(&page); err != nil {
		return nil, err
	}
	return &page, nil
}

func CreateSnapshotCompleted(credentials *config.CloudCredentials, name string, volumeID string) (
	*snapshots.Snapshot, error) {
	opts := &snapshots.CreateOpts{
		VolumeID: volumeID,
		Name:     name,
		Force:    true,
	}
	log.V(4).Infof("[DEBUG] createSnapshot opts: %v", *opts)
	snap, err := CreateSnapshot(credentials, opts)
	if err != nil {
		return nil, err
	}
	log.V(4).Infof("[DEBUG] createSnapshot response detail: %v", snap)
	err = WaitSnapshotReady(credentials, snap.ID)
	if err != nil {
		return nil, err
	}
	return snap, nil
}

func CreateSnapshot(c *config.CloudCredentials, opts *snapshots.CreateOpts) (*snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}
	if !c.Global.Hcs {
		return snapshots.Create(client, opts).Extract()
	}

	b, err := opts.ToSnapshotCreateMap()
	if err != nil {
		return nil, err
	}
	var r snapshots.CreateResult
	_, r.Err = client.Post(client.ServiceURL("snapshots"), b, &r.Body, &golangsdk.RequestOpts{OkCodes: []int{200, 202}})
	return r.Extract()
}

func WaitSnapshotReady(c *config.CloudCredentials, snapshotID string) error {
	availableStatus := "available"
	creatingStatus := "creating"
	err := common.WaitForCompleted(func() (done bool, err error) {
		snapshot, err := GetSnapshot(c, snapshotID)
		if err != nil {
			return false, status.Errorf(codes.Internal,
				"Failed to query snapshot when wait snapshot ready: %v", err)
		}
		log.V(4).Infof("[DEBUG] query snapshot detail when wait snapshot ready detail: %v", snapshot)
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

func DeleteSnapshot(c *config.CloudCredentials, snapshotID string) error {
	client, err := getEvsV2Client(c)
	if err != nil {
		return err
	}
	if !c.Global.Hcs {
		return snapshots.Delete(client, snapshotID).ExtractErr()
	}

	_, err = client.Delete(client.ServiceURL("snapshots", snapshotID), &golangsdk.RequestOpts{OkCodes: []int{200, 202, 204}})
	return err
}
