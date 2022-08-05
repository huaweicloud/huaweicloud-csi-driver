package services

import (
	"fmt"
	"github.com/chnsz/golangsdk/pagination"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"net/url"

	"github.com/chnsz/golangsdk/openstack/evs/v2/snapshots"
)

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

func CreateSnapshot(c *config.CloudCredentials, opts *snapshots.CreateOpts) (*snapshots.Snapshot, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	return snapshots.Create(client, opts).Extract()
}

func WaitSnapshotReady(c *config.CloudCredentials, snapshotId string) error {
	snapshotReadyStatus := "available"
	err := common.WaitForCompleted(func() (done bool, err error) {
		snapshot, err := GetSnapshot(c, snapshotId)
		if err != nil {
			fmt.Printf("WaitSnapshotReady failed to query snapshot: %v", err)
			return false, err
		}
		return snapshot.Status == snapshotReadyStatus, nil
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

	var snaps []snapshots.Snapshot
	err = snapshots.List(client, opts).EachPage(func(page pagination.Page) (bool, error) {
		var err error
		snaps, err = snapshots.ExtractSnapshots(page)
		if err != nil {
			return false, err
		}
		nextPageURL, err := page.NextPageURL()
		if err != nil {
			return false, err
		}
		if nextPageURL != "" {
			queryParams, err := url.ParseQuery(nextPageURL)
			if err != nil {
				return false, err
			}
			fmt.Printf("Query snapshot list param : %v", queryParams)
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return snaps, nil
}
