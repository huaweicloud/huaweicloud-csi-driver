package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"

	"github.com/stretchr/testify/assert"
)

const configFile = "./cloud_config"

var (
	region    = os.Getenv("HW_REGION_NAME")
	accessKey = os.Getenv("HW_ACCESS_KEY")
	secretKey = os.Getenv("HW_SECRET_KEY")
	projectID = os.Getenv("HW_PROJECT_ID")
)

func TestLoadConfig(t *testing.T) {
	err := initConfigFile()
	assert.Nil(t, err)
	defer utils.DeleteFile(configFile)

	cc, err := config.LoadConfig(configFile)
	assert.Nil(t, err)

	assert.EqualValues(t, region, cc.Global.Region)
	assert.EqualValues(t, accessKey, cc.Global.AccessKey)
	assert.EqualValues(t, secretKey, cc.Global.SecretKey)
	assert.EqualValues(t, projectID, cc.Global.ProjectID)
	assert.EqualValues(t, "https://iam.myhuaweicloud.com:443/v3/", cc.Global.AuthURL)
	assert.EqualValues(t, "myhuaweicloud.com", cc.Global.Cloud)
}

func TestEndpoint(t *testing.T) {
	err := initConfigFile()
	assert.Nil(t, err)
	defer utils.DeleteFile(configFile)

	cc, err := config.LoadConfig(configFile)
	assert.Nil(t, err)

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
	expectedURL = fmt.Sprintf("https://ecs.%s.%s/v1/%s/", region, cc.Global.Cloud, client.ProjectID)
	actualURL = client.ResourceBaseURL()
	if actualURL != expectedURL {
		t.Fatalf("ECS endpoint: expected %s but got %s", expectedURL, actualURL)
	}
}

func initConfigFile() error {
	content := fmt.Sprintf(`[Global]
region=%s
access-key=%s
secret-key=%s
project-id=%s
`, region, accessKey, secretKey, projectID)

	err := utils.WriteToFile(configFile, content)
	if err != nil {
		return fmt.Errorf("Error creating cloud config file: %s", err)
	}
	return nil
}
