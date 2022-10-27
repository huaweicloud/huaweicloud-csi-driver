package test

import (
	"fmt"
	"os"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
)

const configFile = "./cloud_config"

var (
	Region       = os.Getenv("HW_REGION_NAME")
	Availability = os.Getenv("HW_AVAILABILITY")
	AccessKey    = os.Getenv("HW_ACCESS_KEY")
	SecretKey    = os.Getenv("HW_SECRET_KEY")
	ProjectID    = os.Getenv("HW_PROJECT_ID")

	NodeID = os.Getenv("HW_NODE_ID")
)

func LoadConfig() (*config.CloudCredentials, error) {
	err := initConfigFile()
	if err != nil {
		return nil, err
	}
	defer utils.DeleteFile(configFile) //nolint:errcheck

	cc, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func initConfigFile() error {
	content := fmt.Sprintf(`[Global]
region=%s
access-key=%s
secret-key=%s
project-id=%s
`, Region, AccessKey, SecretKey, ProjectID)

	err := utils.WriteToFile(configFile, content)
	if err != nil {
		return fmt.Errorf("Error creating cloud config file: %s", err)
	}
	return nil
}
