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

package sfsturbo

import (
	"fmt"
	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/shares"
	"k8s.io/klog"
	"time"
)

const (
	shareAvailable     = "200"

	shareDescription = "provisioned-by=sfsturbo.csi.huaweicloud.org"
)

func createShare(client *golangsdk.ServiceClient, createOpts *shares.CreateOpts) (*shares.TurboResponse, error) {
    createOpts.Description = shareDescription
	share, err := shares.Create(client, createOpts).Extract()
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

// WaitForCreateTurbo create new share from sfsturbo
func waitForShareStatus(client *golangsdk.ServiceClient, shareID string) error {

	retryInterval := 30
	// wait 10 min
	for retryTimes := 0; retryTimes <= 50; retryTimes++ {
		time.Sleep(time.Duration(retryInterval) * time.Second)
		share, err := getShare(client, shareID)
		if err != nil {
			return fmt.Errorf("Failed to get share: %v", err)
		}
		klog.V(2).Infof("create share %s, status is %s", share.ID, share.Status)
		if share.Status == shareAvailable {
			return nil
		}
	}
	return fmt.Errorf("wait share timeout, id: %s", shareID)
}

func getShare(client *golangsdk.ServiceClient, shareID string) (*shares.Turbo, error) {
	return shares.Get(client, shareID).Extract()
}
