package config

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack"
	"github.com/tjfoc/gmsm/gmtls"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common/transport"
)

const (
	UserAgent = "huaweicloud-kubernetes-csi"
)

// CloudCredentials define
type CloudCredentials struct {
	Global struct {
		Cloud     string `gcfg:"cloud"`
		AuthURL   string `gcfg:"auth-url"`
		Region    string `gcfg:"region"`
		Insecure  bool   `gcfg:"insecure"`
		AccessKey string `gcfg:"access-key"`
		SecretKey string `gcfg:"secret-key"`
		ProjectID string `gcfg:"project-id"`
		Idc       bool   `gcfg:"idc"`
		GMSupport bool   `gcfg:"gm-support"`
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
	"ecs": {
		Name:    "ecs",
		Version: "v1",
	},
	"ecsV21": {
		Name:    "ecs",
		Version: "v2.1",
	},
	"evsV1": {
		Name:    "evs",
		Version: "v1",
	},
	"evsV2": {
		Name:    "evs",
		Version: "v2",
	},
	"evsV21": {
		Name:    "evs",
		Version: "v2.1",
	},
	"sfsV2": {
		Name:    "sfs",
		Version: "v2",
	},
	"sfsTurboV1": {
		Name:    "sfs-turbo",
		Version: "v1",
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
	err := c.newCloudClient()
	if err != nil {
		return err
	}
	return nil
}

func (c *CloudCredentials) newCloudClient() error {
	ao := golangsdk.AKSKAuthOptions{
		IdentityEndpoint: c.Global.AuthURL,
		AccessKey:        c.Global.AccessKey,
		SecretKey:        c.Global.SecretKey,
		ProjectId:        c.Global.ProjectID,
		ProjectName:      c.Global.Region,
	}

	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return err
	}

	defaultTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: c.Global.Insecure,
		},
	}

	wrappers := []transport.WrapperFunc{
		transport.NewLogRoundTripper(),
	}
	if c.Global.GMSupport {
		gmCfg := &gmtls.Config{
			GMSupport:          &gmtls.GMSupport{WorkMode: gmtls.ModeAutoSwitch},
			InsecureSkipVerify: c.Global.Insecure,
		}
		wrappers = append(wrappers, transport.NewGMRoundTripper(gmCfg))
	}

	client.HTTPClient = http.Client{
		Transport: transport.Wrappers(wrappers...)(defaultTransport),
	}

	if err := openstack.Authenticate(client, ao); err != nil {
		return err
	}

	c.CloudClient = client
	c.CloudClient.UserAgent.Prepend(UserAgent)

	return nil
}

func (c *CloudCredentials) SFSTurboV1Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "sfsTurboV1", c.Global.Region)
}

func (c *CloudCredentials) SFSV2Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "sfsV2", c.Global.Region)
}

func (c *CloudCredentials) EcsV1Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "ecs", c.Global.Region)
}

func (c *CloudCredentials) EcsV21Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "ecsV21", c.Global.Region)
}

func (c *CloudCredentials) EvsV2Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "evsV2", c.Global.Region)
}

func (c *CloudCredentials) EvsV21Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "evsV21", c.Global.Region)
}

func (c *CloudCredentials) EvsV1Client() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "evsV1", c.Global.Region)
}
