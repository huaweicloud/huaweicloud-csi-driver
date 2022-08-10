package services

import (
	"fmt"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/ecs/v1/block_devices"
	"github.com/chnsz/golangsdk/openstack/ecs/v1/cloudservers"
	"github.com/chnsz/golangsdk/openstack/ecs/v1/jobs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

func GetServer(c *config.CloudCredentials, serverID string) (*cloudservers.CloudServer, error) {
	client, err := getEcsV1Client(c)
	if err != nil {
		return nil, err
	}

	cs, err := cloudservers.Get(client, serverID).Extract()
	if err != nil {
		if common.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "Error, ECS instance %s does not exist", serverID)
		}
		return nil, status.Errorf(codes.Internal, "Error querying ECS instance details: %s", err)
	}
	return cs, nil
}

func AttachVolumeCompleted(c *config.CloudCredentials, serverID, volumeID string) error {
	client, err := getEcsV1Client(c)
	if err != nil {
		return err
	}

	opts := block_devices.AttachOpts{
		ServerId: serverID,
		VolumeId: volumeID,
	}
	log.V(4).Infof("[DEBUG] The option of attaching volume is %#v", opts)

	job, err := block_devices.Attach(client, opts)
	if err != nil {
		log.Errorf("Attach volume failed, serviceID: %s, volumeID: %s, error: %s", serverID, volumeID, err)
		return status.Error(codes.Internal, fmt.Sprintf("Attach volume failed with error %v", err))
	}

	err = waitForVolumeAttached(c, job.ID)
	if err != nil {
		return status.Error(codes.Internal,
			fmt.Sprintf("Failed to wait EVS volume be attached, jobID: %s, error: %v", job.ID, err))
	}
	return nil
}

func DetachVolumeCompleted(c *config.CloudCredentials, serverID, volumeID string) error {
	client, err := getEcsV1Client(c)
	if err != nil {
		return err
	}
	opts := block_devices.DetachOpts{
		ServerId: serverID,
	}
	log.V(4).Infof("[DEBUG] The option of detaching volume is %#v", opts)

	job, err := block_devices.Detach(client, volumeID, opts)
	if err != nil {
		return err
	}
	log.V(4).Infof("[DEBUG] Detach volume job submitted, job ID is %s, job: %#v", job.ID, job)

	err = waitForVolumeDetached(c, job.ID)
	if err != nil {
		return status.Error(codes.Internal,
			fmt.Sprintf("Error waitting for detaching volume with error %s, job ID: %s", err, job.ID))
	}
	log.V(4).Infof("[DEBUG] Volume detached successfully.")
	return nil
}

func waitForVolumeAttached(c *config.CloudCredentials, jobID string) error {
	condition, err := waitForCondition("attaching", c, jobID)
	if err != nil {
		return err
	}
	return common.WaitForCompleted(condition)
}

func waitForVolumeDetached(c *config.CloudCredentials, jobID string) error {
	condition, err := waitForCondition("detaching", c, jobID)
	if err != nil {
		return err
	}
	return common.WaitForCompleted(condition)
}

func waitForCondition(title string, c *config.CloudCredentials, jobID string) (wait.ConditionFunc, error) {
	client, err := getEcsV1Client(c)
	if err != nil {
		return nil, err
	}

	return func() (bool, error) {
		job, err := jobs.Get(client, jobID)
		if err != nil {
			err = status.Error(codes.Internal,
				fmt.Sprintf("Error waiting for the %s job to be complete, job ID: %s", title, jobID))
			return false, err
		}
		if job.Status == "SUCCESS" {
			return true, nil
		}

		if job.Status == "FAIL" {
			return false, status.Error(codes.Internal,
				fmt.Sprintf("Error waiting for the %s job to be complete, job: %#v", title, job))
		}

		return false, nil
	}, nil
}

func getEcsV1Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.EcsV1Client()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed create ECS V1 client: %s", err))
	}
	return client, nil
}
