package config

import (
	"fmt"
	"testing"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
	acceptance "github.com/huaweicloud/huaweicloud-csi-driver/test"
)

func TestLoadConfig(t *testing.T) {
	cc, err := acceptance.LoadConfig()
	if err != nil {
		t.Errorf("Error loading and verifying config data: %s", err)
	}
	assertBasicObj(t, "region", cc.Global.Region, acceptance.Region)
	assertBasicObj(t, "accessKey", cc.Global.AccessKey, acceptance.AccessKey)
	assertBasicObj(t, "secretKey", cc.Global.SecretKey, acceptance.SecretKey)
	assertBasicObj(t, "projectID", cc.Global.ProjectID, acceptance.ProjectID)
	assertBasicObj(t, "authURL", cc.Global.AuthURL, "https://iam.myhuaweicloud.com:443/v3/")
	assertBasicObj(t, "cloud", cc.Global.Cloud, "myhuaweicloud.com")
}

func TestEndpoint(t *testing.T) {
	cc, err := acceptance.LoadConfig()
	if err != nil {
		t.Errorf("Error loading and verifying config data: %s", err)
	}
	err = cc.Validate()
	if err != nil {
		t.Fatal(err)
	}

	var client *golangsdk.ServiceClient
	var expectedURL string
	var actualURL string

	// test ECS client
	client, err = cc.EcsV1Client()
	if err != nil {
		t.Fatalf("Error creating ECS client: %s", err)
	}
	expectedURL = fmt.Sprintf("https://ecs.%s.%s/v1/%s/", acceptance.Region, cc.Global.Cloud, client.ProjectID)
	actualURL = client.ResourceBaseURL()
	if actualURL != expectedURL {
		t.Fatalf("ECS endpoint: expected %s but got %s", expectedURL, actualURL)
	}
}

func assertBasicObj(t *testing.T, name string, a, b interface{}) {
	if a != b {
		t.Errorf("%s expectd: %v, but got: %v", name, a, b)
	}
}

func TestName(t *testing.T) {
	fmt.Println(utils.RandomString(10))
	fmt.Println(utils.RandomString(10))
	fmt.Println(utils.RandomString(10))
	fmt.Println(utils.RandomString(10))
	fmt.Println(utils.RandomString(10))
}
