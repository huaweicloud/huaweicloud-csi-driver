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
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type Bucket struct {
	BucketName          string
	Region              string
	AZRedundancy        string
	EnterpriseProjectID string
	Capacity            int64
}

func GetParallelFSBucket(c *config.CloudCredentials, bucketName string) (*Bucket, error) {
	metadata, err := GetBucketMetadata(c, bucketName)
	if err != nil {
		return nil, err
	}
	if isParallelFile := IsParallelFSBucket(metadata.FSStatus); !isParallelFile {
		return nil, status.Errorf(codes.Unavailable, "Error, the OBS instance %s is not a parallel file system", bucketName)
	}
	capacity, err := GetBucketCapacity(c, bucketName)
	if err != nil {
		return nil, err
	}
	bucket := &Bucket{
		BucketName:          bucketName,
		Region:              metadata.Location,
		AZRedundancy:        metadata.AZRedundancy,
		EnterpriseProjectID: metadata.Epid,
		Capacity:            capacity,
	}
	return bucket, nil
}

func GetBucketMetadata(c *config.CloudCredentials, bucketName string) (*obs.GetBucketMetadataOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.GetBucketMetadataInput{Bucket: bucketName}
	output, err := client.GetBucketMetadata(input)
	if err == nil {
		return output, nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return nil, status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return nil, status.Errorf(codes.Internal, "Error getting OBS instance %s mate data: %v", bucketName, err)
}

func CheckBucketExist(c *config.CloudCredentials, bucketName string) (bool, error) {
	_, err := GetBucketMetadata(c, bucketName)
	if err != nil {
		return false, err
	}
	return true, nil
}

func IsParallelFSBucket(FSStatus obs.FSStatusType) bool {
	return FSStatus == obs.FSStatusEnabled
}

func CreateBucket(c *config.CloudCredentials, bucketName string, acl obs.AclType) error {
	client, err := getObsClient(c)
	if err != nil {
		return err
	}
	input := &obs.CreateBucketInput{
		Bucket:            bucketName,
		ACL:               acl,
		IsFSFileInterface: true,
		BucketLocation:    obs.BucketLocation{Location: c.Global.Region},
	}
	if _, err = client.CreateBucket(input); err == nil {
		return nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusConflict {
		return status.Errorf(codes.AlreadyExists, "Error, the OBS instance %s already exists: %v", bucketName, err)
	}
	return status.Errorf(codes.Internal, "Error creating OBS instance %s: %v", bucketName, err)
}

func CleanBucket(c *config.CloudCredentials, bucketName string) error {
	if err := AbortMultipartUpload(c, bucketName); err != nil {
		return err
	}
	if err := DeleteObjects(c, bucketName); err != nil {
		return err
	}
	return nil
}

func DeleteBucket(c *config.CloudCredentials, bucketName string) error {
	client, err := getObsClient(c)
	if err != nil {
		return err
	}
	_, err = client.DeleteBucket(bucketName)
	if err == nil {
		return nil
	}
	if obsError, ok := err.(obs.ObsError); ok {
		if obsError.StatusCode == http.StatusNotFound {
			return status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
		}
		if obsError.StatusCode == http.StatusConflict {
			return status.Errorf(codes.Unavailable, "Error deleting OBS instance %s is not empty: %v", bucketName, err)
		}
	}
	return status.Errorf(codes.Internal, "Error deleting OBS instance %s: %v", bucketName, err)
}

func AddBucketTags(c *config.CloudCredentials, bucketName string, tags []obs.Tag) error {
	client, err := getObsClient(c)
	if err != nil {
		return err
	}
	input := &obs.SetBucketTaggingInput{
		Bucket:        bucketName,
		BucketTagging: obs.BucketTagging{Tags: tags},
	}
	_, err = client.SetBucketTagging(input)
	if err == nil {
		return nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return status.Errorf(codes.Internal, "Error setting OBS instance %s tag: %v", bucketName, err)
}

func ListBucketTags(c *config.CloudCredentials, bucketName string) ([]obs.Tag, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	output, err := client.GetBucketTagging(bucketName)
	if err == nil {
		return output.Tags, nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		if obsError.Code == "NoSuchTagSet" {
			return []obs.Tag{}, nil
		}
		return nil, status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return nil, status.Errorf(codes.Internal, "Error getting OBS instance %s tags: %v", bucketName, err)
}

func GetBucketStorage(c *config.CloudCredentials, bucketName string) (int64, int, error) {
	client, err := getObsClient(c)
	if err != nil {
		return 0, 0, err
	}
	output, err := client.GetBucketStorageInfo(bucketName)
	if err == nil {
		return output.Size, output.ObjectNumber, nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return 0, 0, status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return 0, 0, status.Errorf(codes.Internal, "Error getting OBS instance %s storage: %v", bucketName, err)
}

func GetBucketCapacity(c *config.CloudCredentials, bucketName string) (int64, error) {
	client, err := getObsClient(c)
	if err != nil {
		return 0, err
	}
	quota, err := client.GetBucketQuota(bucketName)
	if err == nil {
		return quota.Quota, nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return 0, status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return 0, status.Errorf(codes.Internal, "Error getting OBS instance %s capacity: %v", bucketName, err)
}

type ListOpts struct {
	Marker string
	Limit  int
}

func ListBuckets(c *config.CloudCredentials, opts ListOpts) ([]*Bucket, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListBucketsInput{
		QueryLocation: false,
		BucketType:    obs.POSIX,
	}
	output, err := client.ListBuckets(input)
	if err != nil {
		return nil, err
	}

	i := -1
	if opts.Marker != "" {
		i = 0
	}
	for ; opts.Marker != "" && i < len(output.Buckets) && opts.Marker != output.Buckets[i].Name; i++ {
	}

	bucketCount := min(len(output.Buckets)-1-i, opts.Limit)
	bucketList := make([]*Bucket, bucketCount)
	group := errgroup.Group{}
	for k, j := 0, i+1; j <= i+opts.Limit && j < len(output.Buckets); k, j = k+1, j+1 {
		func(bucketName string, idx int) {
			group.Go(func() error {
				fsBucket, err := GetParallelFSBucket(c, bucketName)
				bucketList[idx] = fsBucket
				return err
			})
		}(output.Buckets[j].Name, k)
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return bucketList, nil
}

func SetBucketCapacity(c *config.CloudCredentials, bucketName string, capacity int64) error {
	client, err := getObsClient(c)
	if err != nil {
		return err
	}
	if capacity < 0 || capacity == int64(0) {
		return nil
	}
	input := &obs.SetBucketQuotaInput{
		Bucket: bucketName,
		BucketQuota: obs.BucketQuota{
			Quota: capacity,
		},
	}
	_, err = client.SetBucketQuota(input)
	if err == nil {
		return nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return status.Errorf(codes.Internal, "Error setting OBS instance %s capacity: %v", bucketName, err)
}

func min(i int, j int) int {
	if i < j {
		return i
	}
	return j
}

func getObsClient(c *config.CloudCredentials) (*obs.ObsClient, error) {
	httpClient := c.CloudClient.HTTPClient
	httpClientConfigure := obs.WithHttpClient(&httpClient)

	userAgentConfigure := obs.WithUserAgent(config.UserAgent)

	obsName := "obs"
	if c.Global.Cloud == "prod-cloud-ocb.orange-business.com" {
		obsName = "oss"
	}
	endpoint := fmt.Sprintf("%s.%s.%s", obsName, c.Global.Region, c.Global.Cloud)

	client, err := obs.New(c.Global.AccessKey, c.Global.SecretKey, endpoint, httpClientConfigure, userAgentConfigure)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error initializing OBS client: %v", err)
	}
	if initLog() != nil {
		return nil, status.Errorf(codes.Internal, "Error initializing OBS client log: %v", err)
	}
	return client, nil
}

func initLog() error {
	return obs.InitLog("", 0, 0, obs.LEVEL_INFO, true)
}
