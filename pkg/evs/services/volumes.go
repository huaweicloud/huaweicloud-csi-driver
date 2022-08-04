package services

import (
	"fmt"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/evs/v1/jobs"
	"github.com/chnsz/golangsdk/openstack/evs/v2/cloudvolumes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

func CreateVolumeToCompletion(c *config.CloudCredentials, opts cloudvolumes.CreateOptsBuilder) (string, error) {
	client, err := getEvsV21Client(c)
	if err != nil {
		return "", err
	}

	job, err := cloudvolumes.Create(client, opts).Extract()
	if err != nil {
		return "", fmt.Errorf("error creating EVS volume: %s, options: %#v", err, opts)
	}

	return WaitForCreateEvsJob(c, job.JobID)
}

func GetVolume(c *config.CloudCredentials, id string) (*cloudvolumes.Volume, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	return cloudvolumes.Get(client, id).Extract()
}

func DeleteVolume(c *config.CloudCredentials, id string) error {
	client, err := getEvsV2Client(c)
	if err != nil {
		return err
	}

	return cloudvolumes.Delete(client, id, nil).Err
}

func ListVolumes(c *config.CloudCredentials, opts cloudvolumes.ListOptsBuilder) ([]cloudvolumes.Volume, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	volumes, err := cloudvolumes.ListPage(client, opts)
	if err != nil {
		return nil, fmt.Errorf("error querying a list of the EVS: %s", err)
	}
	return volumes, nil
}

func WaitForCreateEvsJob(c *config.CloudCredentials, jobID string) (string, error) {
	client, err := getJobV1Client(c)
	if err != nil {
		return "", err
	}

	volumeID := ""
	err = common.WaitForCompleted(func() (bool, error) {
		job, err := jobs.GetJobDetails(client, jobID).ExtractJob()
		if err != nil {
			err = status.Error(codes.Internal,
				fmt.Sprintf("error waiting for the creation job to be complete, jobId = %s", jobID))
			return false, err
		}

		if job.Status == "SUCCESS" {
			volumeID = job.Entities.VolumeID
			return true, nil
		}

		if job.Status == "FAIL" {
			err = status.Error(codes.Unavailable, fmt.Sprintf("Error in job creating volume, jobId = %s", jobID))
			return false, err
		}

		return false, nil
	})

	return volumeID, err
}

func getEvsV2Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EvsV2Client()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed create EVS V2 client: %s", err))
	}
	return client, nil
}

func getEvsV21Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EvsV21Client()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed create EVS V2.1 client: %s", err))
	}
	return client, nil
}

func getJobV1Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EvsV1Client()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed create JOB client V1 client: %s", err))
	}
	return client, nil
}
