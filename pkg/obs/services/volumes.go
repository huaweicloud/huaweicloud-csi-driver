package services

import (
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ListObjects(c *config.CloudCredentials, bucket string) (*obs.ListObjectsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListObjectsInput{}
	input.Bucket = bucket
	objects, err := client.ListObjects(input)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Error, OBS instance %s does not exist", bucket)
	}
	return objects, nil
}

func DeleteObjects(c *config.CloudCredentials, bucket string, objects *obs.ListObjectsOutput) (*obs.DeleteObjectsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	// list all object in the bucket
	listOutput, err := ListObjects(c, bucket)
	if err != nil {
		return nil, err
	}
	input := &obs.DeleteObjectsInput{}
	input.Bucket = bucket
	needDel := make([]obs.ObjectToDelete, 0, len(listOutput.Contents))
	for _, object := range listOutput.Contents {
		needDel = append(needDel, obs.ObjectToDelete{Key: object.Key})
	}
	input.Objects = needDel
	// delete all objects int the bucket
	deleteObjects, err := client.DeleteObjects(input)
	if err != nil {
		return nil, err
	}
	return deleteObjects, nil
}

func AbortMultipartUpload(c *config.CloudCredentials, bucket string) (*obs.ListMultipartUploadsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	uploadsOutput, err := ListMultipartUploads(c, bucket)
	if err != nil {
		return nil, err
	}
	for _, job := range uploadsOutput.Uploads {
		input := &obs.AbortMultipartUploadInput{}
		input.Bucket = bucket
		input.UploadId = job.UploadId
		input.Key = job.Key
		_, err := client.AbortMultipartUpload(input)
		if err != nil {
			return nil, err
		}
	}
	return uploadsOutput, nil
}

func ListMultipartUploads(c *config.CloudCredentials, bucket string) (*obs.ListMultipartUploadsOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.ListMultipartUploadsInput{}
	input.Bucket = bucket
	output, err := client.ListMultipartUploads(input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
