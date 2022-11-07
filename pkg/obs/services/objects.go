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
	"net/http"
)

func ListObjects(c *config.CloudCredentials, bucketName string, maxKeys int) (*obs.ListObjectsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListObjectsInput{
		Bucket:        bucketName,
		ListObjsInput: obs.ListObjsInput{MaxKeys: maxKeys},
	}
	objects, err := client.ListObjects(input)
	if err == nil {
		return objects, nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return nil, status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return nil, status.Errorf(codes.Internal, "Error getting OBS instance %s object list", bucketName)
}

func DeleteObjects(c *config.CloudCredentials, bucketName string) error {
	client, err := getObsClient(c)
	if err != nil {
		return err
	}
	for {
		listOutput, err := ListObjects(c, bucketName, 1000)
		if err != nil {
			return err
		}
		objectCount := len(listOutput.Contents)
		if objectCount == 0 {
			break
		}

		objects := make([]obs.ObjectToDelete, 0, objectCount)
		for _, obj := range listOutput.Contents {
			objects = append(objects, obs.ObjectToDelete{Key: obj.Key})
		}

		output, err := client.DeleteObjects(&obs.DeleteObjectsInput{
			Bucket:  bucketName,
			Objects: objects,
		})
		if err != nil {
			return status.Errorf(codes.Internal, "Error deleting OBS instance %s object: %v", bucketName, err)
		}
		if len(output.Deleteds) != objectCount {
			return status.Errorf(codes.Internal, "Error deleting OBS instance %s object, fail num: %d", bucketName, len(output.Errors))
		}
	}
	return nil
}

func AbortMultipartUpload(c *config.CloudCredentials, bucketName string) error {
	client, err := getObsClient(c)
	if err != nil {
		return err
	}
	uploadsOutput, err := ListMultipartUploads(c, bucketName)
	if err != nil {
		return err
	}
	for _, job := range uploadsOutput.Uploads {
		input := &obs.AbortMultipartUploadInput{Bucket: bucketName, UploadId: job.UploadId, Key: job.Key}
		if _, err = client.AbortMultipartUpload(input); err != nil {
			return status.Errorf(codes.Internal, "Error aborting OBS instance %s upload multi upload job: %v", bucketName, err)
		}
	}
	return nil
}

func ListMultipartUploads(c *config.CloudCredentials, bucketName string) (*obs.ListMultipartUploadsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListMultipartUploadsInput{Bucket: bucketName}
	output, err := client.ListMultipartUploads(input)
	if err == nil {
		return output, nil
	}
	if obsError, ok := err.(obs.ObsError); ok && obsError.StatusCode == http.StatusNotFound {
		return nil, status.Errorf(codes.NotFound, "Error, the OBS instance %s does not exist: %v", bucketName, err)
	}
	return output, status.Errorf(codes.Internal, "Error getting OBS instance %s upload list: %v", bucketName, err)
}
