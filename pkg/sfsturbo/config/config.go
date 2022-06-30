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
	"errors"
	"os"

	"github.com/golang/glog"

	"gopkg.in/gcfg.v1"
)

// LoadConfig from file
func LoadConfig(configFile string) (cc CloudCredentials, err error) {
	//Check file path
	if configFile == "" {
		return cc, errors.New("Must provide a config file")
	}

	// Get config from file
	glog.Infof("load config from file: %s", configFile)
	file, err := os.Open(configFile)
	if err != nil {
		return cc, err
	}
	defer file.Close()

	// Read configuration
	err = gcfg.FatalOnly(gcfg.ReadInto(&cc, file))
	if err != nil {
		return cc, err
	}

	// Validate configuration
	err = cc.Validate()
	if err != nil {
		return cc, err
	}

	return cc, nil
}
