/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sfs

import (
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/sfs/v2/shares"
	"k8s.io/klog"
)

const (
	waitForAvailableShareTimeout = 3

	shareAvailable     = "available"

	shareDescription = "provisioned-by=sfs.csi.huaweicloud.org"
)

func createShare(client *golangsdk.ServiceClient, createOpts *shares.CreateOpts) (*shares.Share, error) {
    createOpts.Description = shareDescription
	share, err := shares.Create(client, createOpts).Extract()
	if err != nil {
		return nil, err
	}

	err = waitForShareStatus(client, share.ID, shareAvailable, waitForAvailableShareTimeout)
	if err != nil {
		return nil, err
	}
	return share, nil
}

func deleteShare(client *golangsdk.ServiceClient, shareID string) error {
	if err := shares.Delete(client, shareID).ExtractErr(); err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			klog.V(4).Infof("share %s not found, assuming it to be already deleted", shareID)
		} else {
			return err
		}
	}

	return nil
}

// waitForShareStatus wait for share desired status until timeout
func waitForShareStatus(client *golangsdk.ServiceClient, shareID string, desiredStatus string, timeout int) error {
	return golangsdk.WaitFor(timeout, func() (bool, error) {
		share, err := getShare(client, shareID)
		if err != nil {
			return false, err
		}
		return share.Status == desiredStatus, nil
	})
}

func getShare(client *golangsdk.ServiceClient, shareID string) (*shares.Share, error) {
	return shares.Get(client, shareID).Extract()
}

func grantAccess(client *golangsdk.ServiceClient, shareID string, vpcid string) error {
	// build GrantAccessOpts
	grantAccessOpts := shares.GrantAccessOpts{}
	grantAccessOpts.AccessLevel = "rw"
	grantAccessOpts.AccessType = "cert"
	grantAccessOpts.AccessTo = vpcid

	// grant access
	_, err := shares.GrantAccess(client, shareID, grantAccessOpts).ExtractAccess()
	if err != nil {
		return err
	}
	return nil
}

func expandShare(client *golangsdk.ServiceClient, shareID string, size int) error {
	expandOpts := shares.ExpandOpts{OSExtend: shares.OSExtendOpts{NewSize: size}}
	expand := shares.Expand(client, shareID, expandOpts)
	if expand.Err != nil {
		return expand.Err
	}
	return nil
}
