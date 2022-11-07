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
	"errors"
	"fmt"
	"os"

	"github.com/golang/glog"

	"gopkg.in/gcfg.v1"
)

// LoadConfig from file
func LoadConfig(configFile string) (*CloudCredentials, error) {
	// Check file path
	if configFile == "" {
		return nil, errors.New("Must provide a config file")
	}

	// Get config from file
	glog.Infof("load config from file: %s", configFile)
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cc := &CloudCredentials{}
	// Read configuration
	err = gcfg.FatalOnly(gcfg.ReadInto(cc, file))
	if err != nil {
		return nil, err
	}

	// Set default value
	setDefaultConfig(cc)

	// Validate configuration
	err = cc.Validate()
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func setDefaultConfig(cc *CloudCredentials) {
	if cc.Global.Cloud == "" {
		cc.Global.Cloud = "myhuaweicloud.com"
	}
	if cc.Global.AuthURL == "" {
		cc.Global.AuthURL = fmt.Sprintf("https://iam.%s:443/v3/", cc.Global.Cloud)
	}
}
