package services

import (
	"fmt"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/evs/v1/jobs"
	"github.com/chnsz/golangsdk/openstack/evs/v2/cloudvolumes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

const (
	EvsAvailableStatus = "available"
)

func CreateVolumeCompleted(c *config.CloudCredentials, otps *cloudvolumes.CreateOpts) (string, error) {
	client, err := getEvsV21Client(c)
	if err != nil {
		return "", err
	}

	job, err := cloudvolumes.Create(client, *otps).Extract()
	if err != nil {
		return "", fmt.Errorf("Error creating EVS Volume, error: %s, createOpts: %#v", err, otps)
	}

	log.V(4).Infof("[DEBUG] The volume creation is submitted successfully and the job is running.")
	return waitForJobFinished(c, "creation", job.JobID)
}

func GetVolume(c *config.CloudCredentials, id string) (*cloudvolumes.Volume, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}

	volume, err := cloudvolumes.Get(client, id).Extract()
	if err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "Error, volume %s does not exist", id)
		}
		return nil, status.Errorf(codes.Internal, "Error querying volume details: %s", err)
	}
	return volume, nil
}

func CheckVolumeExists(credentials *config.CloudCredentials, name string, sizeGB int) (*cloudvolumes.Volume, error) {
	opts := cloudvolumes.ListOpts{
		Name: name,
	}
	volumes, err := ListVolumes(credentials, opts)
	if err != nil {
		return nil, status.Error(codes.Internal,
			fmt.Sprintf("Failed to query the volume by name, cannot verify whether it exists: %s", err))
	}

	if len(volumes) == 1 {
		vol := volumes[0]
		if sizeGB != vol.Size {
			return nil, status.Error(codes.AlreadyExists,
				"A volume already exists with the same name but a different capacity")
		}
		log.Infof("Volume %s already exists in AZ %s of size %d GiB", vol.ID, vol.AvailabilityZone, vol.Size)
		return &vol, nil
	} else if len(volumes) > 1 {
		log.Infof("found multiple existing volumes with selected name (%s) during create", name)
		return nil, status.Error(codes.AlreadyExists, "Found multiple volumes with same name")
	}

	return nil, nil
}

func GetVolumeDevicePath(c *config.CloudCredentials, id string) (string, error) {
	volume, err := GetVolume(c, id)
	log.V(4).Infof("[DEBUG] Got volume details: %#v", volume)

	if err != nil {
		if common.IsNotFound(err) {
			return "", status.Errorf(codes.NotFound, "Volume %s does not exist, get device path failed", id)
		}
		return "", status.Error(codes.Internal, fmt.Sprintf("Querying volume details fails with error %s", err))
	}

	if len(volume.Attachments) > 0 {
		return volume.Attachments[0].Device, nil
	}
	return "", status.Error(codes.Internal, "Get volume device path failed, volume has no attachment")
}

func ExpandVolume(c *config.CloudCredentials, id string, newSize int) error {
	client, err := getEvsV21Client(c)
	if err != nil {
		return err
	}

	opt := cloudvolumes.ExtendOpts{
		SizeOpts: cloudvolumes.ExtendSizeOpts{
			NewSize: newSize,
		},
	}
	log.V(4).Infof("[DEBUG] Expand volume %s, and the options is %#v", id, opt)

	job, err := cloudvolumes.ExtendSize(client, id, opt).Extract()
	if err != nil {
		return status.Error(codes.Internal,
			fmt.Sprintf("Error expanding volume, volume: %s, newSize: %v", id, newSize))
	}
	log.V(4).Infof("[DEBUG] The volume expanding is submitted successfully and the job is running.")

	_, err = waitForJobFinished(c, "expanding", job.JobID)
	return err
}

func DeleteVolume(c *config.CloudCredentials, id string) error {
	client, err := getEvsV2Client(c)
	if err != nil {
		return err
	}

	return cloudvolumes.Delete(client, id, nil).Err
}

func ListVolumes(c *config.CloudCredentials, opts cloudvolumes.ListOpts) ([]cloudvolumes.Volume, error) {
	client, err := getEvsV2Client(c)
	if err != nil {
		return nil, err
	}
	log.V(4).Infof("[DEBUG] Query a volume list, and the options is %#v", opts)

	volumes, err := cloudvolumes.ListPage(client, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error querying volume list, error: %v", err))
	}
	return volumes, nil
}

func waitForJobFinished(c *config.CloudCredentials, title, jobID string) (string, error) {
	client, err := getJobV1Client(c)
	if err != nil {
		return "", err
	}

	var volumeID string
	err = common.WaitForCompleted(func() (bool, error) {
		job, err := jobs.GetJobDetails(client, jobID).ExtractJob()
		if err != nil {
			return false, status.Error(codes.Internal,
				fmt.Sprintf("Error waiting for the %s volume job to be complete, jobID: %s", title, jobID))
		}

		if job.Status == "SUCCESS" {
			volumeID = job.Entities.VolumeID
			return true, nil
		}

		if job.Status == "FAIL" {
			return false, status.Error(codes.Unavailable,
				fmt.Sprintf("Error waiting for the %s volume job to be complete, job: %#v", title, job))
		}

		return false, nil
	})

	return volumeID, err
}

func getEvsV2Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EvsV2Client()
	if err != nil {
		logMsg := fmt.Sprintf("Failed create EVS V2 client: %s", err)
		return nil, status.Error(codes.Internal, logMsg)
	}
	return client, nil
}

func getEvsV21Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EvsV21Client()
	if err != nil {
		logMsg := fmt.Sprintf("Failed create EVS V2.1 client: %s", err)
		return nil, status.Error(codes.Internal, logMsg)
	}
	return client, nil
}

func getJobV1Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EvsV1Client()
	if err != nil {
		logMsg := fmt.Sprintf("Failed create JOB V1 client: %s", err)
		return nil, status.Error(codes.Internal, logMsg)
	}
	return client, nil
}
