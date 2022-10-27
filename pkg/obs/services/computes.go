package services

import (
	"fmt"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetBucket(c *config.CloudCredentials, bucket string) (string, error) {
	exists, err := CheckBucketExists(c, bucket)
	if err != nil || !exists {
		return "", err
	}
	if enabled, e := CheckBucketFSStatus(c, bucket); !enabled {
		if e != nil {
			return "", e
		}
		return "", status.Errorf(codes.Unavailable, "Error, OBS instance %s is not FSFileInterface", bucket)
	}
	return bucket, nil
}

func GetBucketMetadata(c *config.CloudCredentials, bucket string) (*obs.GetBucketMetadataOutput, error) {
	client, err := getObsClient(c)
	if err != nil {
		return nil, err
	}
	input := &obs.GetBucketMetadataInput{}
	input.Bucket = bucket
	metadata, err := client.GetBucketMetadata(input)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Error, OBS instance %s does not exist, error: %v", bucket, err)
	}
	return metadata, nil
}

func CheckBucketExists(c *config.CloudCredentials, bucket string) (bool, error) {
	client, err := getObsClient(c)
	if err != nil {
		return false, err
	}
	_, err = client.HeadBucket(bucket)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			if obsError.StatusCode == 404 {
				return false, nil
			}
		}
		return false, status.Errorf(codes.Internal, "Error, heading OBS instance %s, error:%v", bucket, err)
	}
	return true, nil
}

func CheckBucketFSStatus(c *config.CloudCredentials, bucket string) (bool, error) {
	metadata, err := GetBucketMetadata(c, bucket)
	if err != nil {
		return false, err
	}
	return string(metadata.FSStatus) == "Enabled", nil
}

func CreateBucket(c *config.CloudCredentials, bucket string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	input := &obs.CreateBucketInput{}
	input.Bucket = bucket
	input.Location = c.Global.Region
	input.ACL = obs.AclPrivate
	input.IsFSFileInterface = true
	output, err := client.CreateBucket(input)
	if err != nil {
		if output.StatusCode == 409 {
			return "", status.Errorf(codes.AlreadyExists, "Error, creating OBS instance %s existed, error:%v", bucket, err)
		}
		return "", status.Errorf(codes.Internal, "Error, creating OBS instance %s, error:%v", bucket, err)
	}
	return bucket, nil
}

func CreateBucketWithTag(c *config.CloudCredentials, bucket string) (string, error) {
	_, err := CreateBucket(c, bucket)
	if err != nil {
		return "", err
	}
	_, err = SetBucketCSITagging(c, bucket)
	if err != nil {
		return "", err
	}
	return bucket, nil
}

func DeleteBucket(c *config.CloudCredentials, bucket string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	_, err = client.DeleteBucket(bucket)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Error, deleting OBS instance %s, error:%v", bucket, err)
	}
	return bucket, nil
}

func SetBucketCSITagging(c *config.CloudCredentials, bucket string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	input := &obs.SetBucketTaggingInput{}
	input.Bucket = bucket
	var tags [1]obs.Tag
	tags[0] = obs.Tag{Key: "CSI", Value: "CSI-CREATE"}
	_, err = client.SetBucketTagging(input)
	if err != nil {
		return "", err
	}
	return "CSI", nil
}

func GetBucketCSITagging(c *config.CloudCredentials, bucket string) (string, error) {
	client, err := getObsClient(c)
	if err != nil {
		return "", err
	}
	output, err := client.GetBucketTagging(bucket)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Error, getting OBS instance %s tag, error:%v", bucket, err)
	}
	tag := output.Tags[0].Value
	return tag, nil
}

func GetBucketStorage(c *config.CloudCredentials, bucket string) (int64, int, error) {
	client, err := getObsClient(c)
	if err != nil {
		return 0, 0, err
	}
	output, err := client.GetBucketStorageInfo(bucket)
	if err != nil {
		return 0, 0, status.Errorf(codes.Internal, "Error, getting OBS instance %s storage info, error:%v", bucket, err)
	}
	return output.Size, output.ObjectNumber, nil
}

func GetBucketQuota(c *config.CloudCredentials, bucket string) (int64, error) {
	client, err := getObsClient(c)
	if err != nil {
		return 0, err
	}
	quota, err := client.GetBucketQuota(bucket)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "Error, getting OBS instance %s quota info, error:%v", bucket, err)
	}
	return quota.Quota, nil
}

func getObsClient(c *config.CloudCredentials) (*obs.ObsClient, error) {
	client, err := obs.New(c.Global.AccessKey, c.Global.SecretKey, "obs.cn-north-4.myhuaweicloud.com")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed create OBS client: %s", err))
	}
	return client, nil
}
