package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack"
	"k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
)

const (
	UserAgent = "huaweicloud-kubernetes-csi"

	// retries after a 429; with backoff from 2s to 60s they span about 5 minutes
	defaultAPIRetries = 8
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
		// Hcs switches EVS to the APIs available on Huawei Cloud Stack
		Hcs bool `gcfg:"hcs"`
		// SkipCatalog skips GET /v3/auth/catalog at startup, needs project-id
		SkipCatalog bool `gcfg:"skip-catalog"`
		// CAFile is a PEM bundle trusted in addition to the system CAs
		CAFile string `gcfg:"ca-file"`
		// APIRateLimit caps API calls per second from this process, 0: no limit
		APIRateLimit float64 `gcfg:"api-rate-limit"`
		// APIBurst is how many calls may go out at once under api-rate-limit
		APIBurst int `gcfg:"api-burst"`
		// APIRetries is how often a call rejected with 429 is retried, 0: default, -1: never
		APIRetries int `gcfg:"api-retries"`
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
	"bss": {
		Name:             "bss",
		Version:          "v2",
		Scope:            "global",
		WithOutProjectID: true,
	},
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
		WithUserCatalog:  c.Global.SkipCatalog,
	}
	if c.Global.SkipCatalog && c.Global.ProjectID == "" {
		return fmt.Errorf("skip-catalog requires project-id")
	}

	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return err
	}

	tlsConfig, err := c.tlsConfig()
	if err != nil {
		return err
	}
	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
	}

	retries := c.Global.APIRetries
	if retries == 0 {
		retries = defaultAPIRetries
	} else if retries < 0 {
		retries = 0
	}
	client.HTTPClient = http.Client{
		Transport: utils.NewThrottleRoundTripper(
			&utils.LogRoundTripper{Rt: transport},
			c.Global.APIRateLimit, c.Global.APIBurst, retries),
	}

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return err
	}

	c.CloudClient = client
	c.CloudClient.UserAgent.Prepend(UserAgent)
	return nil
}

func (c *CloudCredentials) tlsConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: c.Global.Insecure,
	}
	if c.Global.CAFile == "" {
		return cfg, nil
	}
	if c.Global.Insecure {
		klog.Warningf("ca-file %s is ignored because insecure is true", c.Global.CAFile)
		return cfg, nil
	}

	pem, err := os.ReadFile(c.Global.CAFile)
	if err != nil {
		return nil, fmt.Errorf("reading ca-file: %v", err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("ca-file %s holds no PEM certificates", c.Global.CAFile)
	}
	cfg.RootCAs = pool
	return cfg, nil
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

func (c *CloudCredentials) BssClient() (*golangsdk.ServiceClient, error) {
	return newServiceClient(c, "bss", c.Global.Region)
}
