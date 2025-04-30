/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	"net/http"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common/transport"
)

const (
	UserAgent = "huaweicloud-kubernetes-csi"
)

// CloudCredentials define
type CloudCredentials struct {
	Global struct {
		Cloud      string `gcfg:"cloud"`
		AuthURL    string `gcfg:"auth-url"`
		Region     string `gcfg:"region"`
		DomainName string `gcfg:"domain-name"`
		Username   string `gcfg:"username"`
		Password   string `gcfg:"password"`
		AccessKey  string `gcfg:"access-key"`
		SecretKey  string `gcfg:"secret-key"`
		ProjectID  string `gcfg:"project-id"`
	}

	Vpc struct {
		ID              string `gcfg:"id"`
		SubnetID        string `gcfg:"subnet-id"`
		SecurityGroupID string `gcfg:"security-group-id"`
	}

	CloudClient *golangsdk.ProviderClient
}

type serviceCatalog struct {
	Name             string
	Version          string
	Scope            string
	Admin            bool
	ResourceBase     string
	WithOutProjectID bool
}

var allServiceCatalog = map[string]serviceCatalog{
	"sfsV2": {
		Name:    "sfs",
		Version: "v2",
	},
}

func newServiceClient(cc *CloudCredentials, catalogName, region string) (*golangsdk.ServiceClient, error) {
	catalog, ok := allServiceCatalog[catalogName]
	if !ok {
		return nil, fmt.Errorf("service type %s is invalid or not supportted", catalogName)
	}

	client := cc.CloudClient
	// update ProjectID and region in ProviderClient
	clone := new(golangsdk.ProviderClient)
	*clone = *client
	clone.ProjectID = client.ProjectID
	clone.AKSKAuthOptions.ProjectId = client.ProjectID
	clone.AKSKAuthOptions.Region = region

	sc := &golangsdk.ServiceClient{
		ProviderClient: clone,
	}

	if catalog.Scope == "global" {
		sc.Endpoint = fmt.Sprintf("https://%s.%s/", catalog.Name, cc.Global.Cloud)
	} else {
		sc.Endpoint = fmt.Sprintf("https://%s.%s.%s/", catalog.Name, region, cc.Global.Cloud)
	}

	sc.ResourceBase = sc.Endpoint
	if catalog.Version != "" {
		sc.ResourceBase = sc.ResourceBase + catalog.Version + "/"
	}
	if !catalog.WithOutProjectID {
		sc.ResourceBase = sc.ResourceBase + client.ProjectID + "/"
	}
	if catalog.ResourceBase != "" {
		sc.ResourceBase = sc.ResourceBase + catalog.ResourceBase + "/"
	}

	return sc, nil
}

func (c *CloudCredentials) Validate() error {
	var authOpt golangsdk.AuthOptionsProvider
	if c.Global.AccessKey != "" && c.Global.SecretKey != "" {
		authOpt = golangsdk.AKSKAuthOptions{
			IdentityEndpoint: c.Global.AuthURL,
			AccessKey:        c.Global.AccessKey,
			SecretKey:        c.Global.SecretKey,
			ProjectId:        c.Global.ProjectID,
			ProjectName:      c.Global.Region,
		}
	} else {
		authOpt = golangsdk.AuthOptions{
			IdentityEndpoint: c.Global.AuthURL,
			TenantName:       c.Global.Region,
			DomainName:       c.Global.DomainName,
			Username:         c.Global.Username,
			Password:         c.Global.Password,
			TenantID:         c.Global.ProjectID,
			AllowReauth:      true,
		}
	}

	client, err := openstack.NewClient(c.Global.AuthURL)
	if err != nil {
		return err
	}

	newTransport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	client.HTTPClient = http.Client{
		Transport: &transport.LogRoundTripper{
			Rt: newTransport,
		},
	}

	err = openstack.Authenticate(client, authOpt)
	if err != nil {
		return err
	}

	c.CloudClient = client
	c.CloudClient.UserAgent.Prepend(UserAgent)
	return nil
}

func (c *CloudCredentials) SFSV2Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "sfsV2", c.Global.Region)
}
