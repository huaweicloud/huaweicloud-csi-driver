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

package config

import (
	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack"
	"k8s.io/klog"
	"net/http"
	"os"
)

// CloudCredentials define
type CloudCredentials struct {
	Global struct {
		TenantID         string `gcfg:"tenant-id"`
		DomainName       string `gcfg:"domain-name"`
		Username         string
		Password         string
		AccessKey        string `gcfg:"access-key"`
		SecretKey        string `gcfg:"secret-key"`
		Region           string `gcfg:"region"`
		AvailabilityZone string `gcfg:"availability-zone"`
		ProjectName      string `gcfg:"project-name"`
		ProjectId        string `gcfg:"project-id"`
		AuthURL          string `gcfg:"auth-url"`
	}

	Vpc struct {
		Id              string `gcfg:"id"`
		SubnetId        string `gcfg:"subnet-id"`
		SecurityGroupId string `gcfg:"security-group-id"`
	}

	Ext struct {
		ShareProto string `gcfg:"share-proto"`
	}

	CloudClient *golangsdk.ProviderClient
}

// Validate CloudCredentials
func (c *CloudCredentials) Validate() error {
	err := c.newCloudClient()
	if err != nil {
		return err
	}
	return nil
}

// newCloudClient returns new cloud client
func (c *CloudCredentials) newCloudClient() error {
	ao := golangsdk.AKSKAuthOptions{
		IdentityEndpoint: c.Global.AuthURL,
		AccessKey:        c.Global.AccessKey,
		SecretKey:        c.Global.SecretKey,
		Region:           c.Global.Region,
		ProjectName:      c.Global.ProjectName,
		ProjectId:        c.Global.ProjectId,
	}

	passAo := golangsdk.AuthOptions{
		IdentityEndpoint: c.Global.AuthURL,
		Username:         c.Global.Username,
		Password:         c.Global.Password,
		DomainName:       c.Global.DomainName,
		TenantID:         c.Global.TenantID,
		AllowReauth:      true,
	}

	var client *golangsdk.ProviderClient
	var err error

	if c.Global.AccessKey == "" || c.Global.SecretKey == "" {
		klog.V(4).Infof("ak sk client")
		client, err = openstack.NewClient(ao.IdentityEndpoint)
	} else {
		klog.V(4).Infof("password client")
		client, err = openstack.NewClient(passAo.IdentityEndpoint)
	}
	if err != nil {
		return err
	}

	// if OS_DEBUG is set, log the requests and responses
	var osDebug bool
	if os.Getenv("OS_DEBUG") != "" {
		osDebug = true
	}

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	client.HTTPClient = http.Client{
		Transport: &LogRoundTripper{
			Rt:      transport,
			OsDebug: osDebug,
		},
	}

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return err
	}

	c.CloudClient = client

	return nil
}

func (c *CloudCredentials) SFSTurboV1Client() (*golangsdk.ServiceClient, error) {
	return openstack.NewHwSFSTurboV1(c.CloudClient, golangsdk.EndpointOpts{
		Region:       c.Global.Region,
		Availability: golangsdk.AvailabilityPublic,
	})
}
