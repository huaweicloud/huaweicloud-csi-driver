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

package services

import (
	"fmt"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func GetBucket(c *config.CloudCredentials, bucketName string) (string, error) {
	exists, err := CheckBucketExists(c, bucketName)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	isParallelFile, err := CheckBucketFSStatus(c, bucketName)
	if err != nil {
		return "", err
	}
	if !isParallelFile {
		return "", status.Errorf(codes.Unavailable, "Error, the OBS bucket %s is not a parallel file system", bucketName)
	}
	return bucketName, nil
}

func GetBucketMetadata(c *config.CloudCredentials, bucketName string) (*obs.GetBucketMetadataOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.GetBucketMetadataInput{}
	input.Bucket = bucketName
	output, err := client.GetBucketMetadata(input)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Error, OBS instance %s does not exist, error: %v", bucketName, err)
	}
	return output, nil
}

func CheckBucketExists(c *config.CloudCredentials, bucketName string) (bool, error) {
	client, err := getObsClient(c)
	if err != nil {
		return false, err
	}
	_, err = client.HeadBucket(bucketName)
	if err == nil {
		return true, nil
	}
	if obsError, ok := err.(obs.ObsError); ok {
		if obsError.StatusCode == http.StatusNotFound {
			return false, nil
		}
	}
	return false, status.Errorf(codes.Internal, "Error, heading OBS instance %s, error: %v", bucketName, err)
}

func CheckBucketFSStatus(c *config.CloudCredentials, bucketName string) (bool, error) {
	metadata, err := GetBucketMetadata(c, bucketName)
	if err != nil {
		return false, err
	}
	return metadata.FSStatus == obs.FSStatusEnabled, nil
}

func CreateBucket(c *config.CloudCredentials, bucketName string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	input := &obs.CreateBucketInput{}
	input.Bucket = bucketName
	input.Location = c.Global.Region
	input.ACL = obs.AclPrivate
	input.IsFSFileInterface = true
	_, err = client.CreateBucket(input)
	if err == nil {
		return bucketName, nil
	}
	if obsError, ok := err.(obs.ObsError); ok {
		if obsError.StatusCode == http.StatusConflict {
			return "", status.Errorf(codes.AlreadyExists, "Error, creating OBS instance %s existed, error: %v", bucketName, err)
		}
	}
	return "", status.Errorf(codes.Internal, "Error, creating OBS instance %s, error: %v", bucketName, err)
}

func CreateBucketWithTag(c *config.CloudCredentials, bucketName string) (string, error) {
	_, err := CreateBucket(c, bucketName)
	if err != nil {
		return "", err
	}
	_, err = SetBucketCSITagging(c, bucketName)
	if err != nil {
		return "", err
	}
	return bucketName, nil
}

func DeleteBucket(c *config.CloudCredentials, bucketName string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	_, err = client.DeleteBucket(bucketName)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Error, deleting OBS instance %s, error: %v", bucketName, err)
	}
	return bucketName, nil
}

func SetBucketCSITagging(c *config.CloudCredentials, bucketName string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	input := &obs.SetBucketTaggingInput{}
	input.Bucket = bucketName
	input.Tags = []obs.Tag{{Key: "CSI", Value: "CSI-CREATE"}}
	_, err = client.SetBucketTagging(input)
	if err != nil {
		return "", err
	}
	return "CSI", nil
}

func GetBucketCSITagging(c *config.CloudCredentials, bucketName string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	output, err := client.GetBucketTagging(bucketName)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Error, getting OBS instance %s tag, error: %v", bucketName, err)
	}
	tag := output.Tags[0].Value
	return tag, nil
}

func GetBucketStorage(c *config.CloudCredentials, bucketName string) (int64, int, error) {
	client, err := getObsClient(c)
	if err != nil {
		return 0, 0, err
	}
	output, err := client.GetBucketStorageInfo(bucketName)
	if err != nil {
		return 0, 0, status.Errorf(codes.Internal, "Error, getting OBS instance %s storage info, error: %v", bucketName, err)
	}
	return output.Size, output.ObjectNumber, nil
}

func GetBucketQuota(c *config.CloudCredentials, bucketName string) (int64, error) {
	client, err := getObsClient(c)
	if err != nil {
		return 0, err
	}
	quota, err := client.GetBucketQuota(bucketName)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "Error, getting OBS instance %s quota info, error: %v", bucketName, err)
	}
	return quota.Quota, nil
}

func getObsClient(c *config.CloudCredentials) (*obs.ObsClient, error) {
	endpoint := fmt.Sprintf("obs.%s.%s", c.Global.Region, c.Global.Cloud)
	client, err := obs.New(c.Global.AccessKey, c.Global.SecretKey, endpoint)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error, initializing OBS client, error: %v", err)
	}
	if initLog() != nil {
		return nil, status.Errorf(codes.Internal, "Error, initializing OBS client log, error: %v", err)
	}
	return client, nil
}

func initLog() error {
	var logFullPath string = "./logs/OBS-SDK.log"
	var maxLogSize int64 = 1024 * 1024 * 10
	err := obs.InitLog(logFullPath, maxLogSize, 10, obs.LEVEL_INFO, true)
	return err
}
