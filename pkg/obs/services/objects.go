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
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ListObjects(c *config.CloudCredentials, bucketName string) (*obs.ListObjectsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListObjectsInput{}
	input.Bucket = bucketName
	objects, err := client.ListObjects(input)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Error, OBS instance %s does not exist", bucketName)
	}
	return objects, nil
}

func DeleteObjects(c *config.CloudCredentials, bucketName string, objects *obs.ListObjectsOutput) (*obs.DeleteObjectsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	// list all object in the bucketName
	listOutput, err := ListObjects(c, bucketName)
	if err != nil {
		return nil, err
	}
	input := &obs.DeleteObjectsInput{}
	input.Bucket = bucketName
	needDel := make([]obs.ObjectToDelete, 0, len(listOutput.Contents))
	for _, object := range listOutput.Contents {
		needDel = append(needDel, obs.ObjectToDelete{Key: object.Key})
	}
	input.Objects = needDel
	// delete all objects int the bucketName
	deleteObjects, err := client.DeleteObjects(input)
	if err != nil {
		return nil, err
	}
	return deleteObjects, nil
}

func AbortMultipartUpload(c *config.CloudCredentials, bucketName string) (*obs.ListMultipartUploadsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	uploadsOutput, err := ListMultipartUploads(c, bucketName)
	if err != nil {
		return nil, err
	}
	for _, job := range uploadsOutput.Uploads {
		input := &obs.AbortMultipartUploadInput{}
		input.Bucket = bucketName
		input.UploadId = job.UploadId
		input.Key = job.Key
		_, err := client.AbortMultipartUpload(input)
		if err != nil {
			return nil, err
		}
	}
	return uploadsOutput, nil
}

func ListMultipartUploads(c *config.CloudCredentials, bucketName string) (*obs.ListMultipartUploadsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListMultipartUploadsInput{}
	input.Bucket = bucketName
	output, err := client.ListMultipartUploads(input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
