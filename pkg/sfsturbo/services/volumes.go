package services

import (
	"strconv"
	"time"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/sfs_turbo/v1/shares"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

const (
	shareCreating  = "100"
	shareAvailable = "200"

	shareSubExpanding     = "121"
	shareSubExpandSuccess = "221"
	shareSubExpandError   = "321"

	shareDescription = "provisioned-by=sfsturbo.csi.huaweicloud.org"

	DefaultInitDelay = 15 * time.Second
	DefaultFactor    = 1.02
	DefaultSteps     = 30
)

func CreateShareCompleted(c *config.CloudCredentials, createOpts *shares.Share) (
	*shares.TurboResponse, error) {
	turboResponse, err := CreateShare(c, &shares.CreateOpts{Share: *createOpts})
	if err != nil {
		return nil, err
	}
	log.V(4).Infof("[DEBUG] create share response detail: %v", protosanitizer.StripSecrets(turboResponse))
	err = WaitForShareAvailable(c, turboResponse.ID)
	if err != nil {
		return nil, err
	}
	return turboResponse, nil
}

// WaitForShareAvailable will cost few minutes
func WaitForShareAvailable(c *config.CloudCredentials, shareID string) error {
	condition := func() (bool, error) {
		share, err := GetShare(c, shareID)
		if err != nil {
			return false, status.Errorf(codes.Internal,
				"Failed to query share %s when wait share available: %v", shareID, err)
		}
		log.V(4).Infof("[DEBUG] WaitForShareAvailable query detail: %v", protosanitizer.StripSecrets(share))
		if share.Status == shareAvailable {
			return true, nil
		}
		if share.Status == shareCreating {
			return false, nil
		}
		return false, status.Errorf(codes.Internal, "created share status is not available : %s", share.Status)
	}
	backoff := wait.Backoff{
		Duration: DefaultInitDelay,
		Factor:   DefaultFactor,
		Steps:    DefaultSteps,
	}
	return wait.ExponentialBackoff(backoff, condition)
}

func CreateShare(c *config.CloudCredentials, createOpts *shares.CreateOpts) (*shares.TurboResponse, error) {
	createOpts.Share.Description = shareDescription
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return nil, err
	}
	share, err := shares.Create(client, createOpts).Extract()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create share, err: %v", err)
	}
	return share, nil
}

func DeleteShareCompleted(c *config.CloudCredentials, shareID string) error {
	if _, err := GetShare(c, shareID); err != nil {
		if common.IsNotFound(err) {
			log.V(4).Infof("[DEBUG] share %s not found, it has already been deleted", shareID)
			return nil
		}
		return status.Errorf(codes.Internal, "Failed to query volume %s, error: %v", shareID, err)
	}
	if err := DeleteShare(c, shareID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete share, err: %v", err)
	}
	return WaitForShareDeleted(c, shareID)
}

func DeleteShare(c *config.CloudCredentials, shareID string) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}
	return shares.Delete(client, shareID).ExtractErr()
}

func WaitForShareDeleted(c *config.CloudCredentials, shareID string) error {
	return common.WaitForCompleted(func() (bool, error) {
		if _, err := GetShare(c, shareID); err != nil {
			if common.IsNotFound(err) {
				// resource not exist
				return true, nil
			}
			return false, status.Errorf(codes.Internal,
				"Failed to query share %s when wait share deleted: %v", shareID, err)
		}
		return false, nil
	})
}

func GetShare(c *config.CloudCredentials, shareID string) (*shares.Turbo, error) {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return nil, err
	}
	return shares.Get(client, shareID).Extract()
}

func ListTotalShares(c *config.CloudCredentials) ([]shares.Turbo, error) {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return nil, err
	}

	page, err := shares.List(client)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to query list total shares: %v", err)
	}
	return page, nil
}

func ListPageShares(c *config.CloudCredentials, opts shares.ListOpts) (*shares.PagedList, error) {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return nil, err
	}

	pageList, err := shares.ListPage(client, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to query list page shares: %v", err)
	}
	log.V(4).Infof("[DEBUG] Query list page shares detail: %v", pageList)
	return pageList, nil
}

func ExpandShareCompleted(c *config.CloudCredentials, id string, newSize int) error {
	if err := ExpandShare(c, id, newSize); err != nil {
		return err
	}
	if err := WaitForShareExpanded(c, id); err != nil {
		return err
	}
	share, err := GetShare(c, id)
	if err != nil {
		return err
	}
	shareSize, err := strconv.ParseFloat(share.Size, 64)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to convert string size to number size")
	}
	if int(shareSize) != newSize {
		return status.Errorf(codes.Internal, "Failed to expand share size, cause get an unexpected volume size")
	}
	return nil
}

func ExpandShare(c *config.CloudCredentials, id string, newSize int) error {
	client, err := getSFSTurboV1Client(c)
	if err != nil {
		return err
	}

	opt := shares.ExpandOpts{
		Extend: shares.ExtendOpts{
			NewSize: newSize,
		},
	}
	log.V(4).Infof("[DEBUG] Expand volume %s, and the options is %#v", id, opt)
	job, err := shares.Expand(client, id, opt).Extract()
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to expand volume: %s, err: %v", id, err)
	}
	log.V(4).Infof("[DEBUG] The volume expanding is submitted successfully and the job is running, job: %v", job)
	return nil
}

func WaitForShareExpanded(c *config.CloudCredentials, shareID string) error {
	return common.WaitForCompleted(func() (bool, error) {
		share, err := GetShare(c, shareID)
		if err != nil {
			return false, status.Errorf(codes.Internal,
				"Failed to query share %s when wait share expand: %v", shareID, err)
		}
		log.V(4).Infof("[DEBUG] query share detail: %v", protosanitizer.StripSecrets(share))
		if share.SubStatus == shareSubExpandSuccess {
			return true, nil
		}
		if share.SubStatus == shareSubExpanding {
			return false, nil
		}
		if share.SubStatus == shareSubExpandError {
			return false, status.Errorf(codes.Internal, "Failed to expand share, cause sub status error")
		}
		return false, status.Error(codes.Internal, "expand share sub status is not success")
	})
}

func getSFSTurboV1Client(c *config.CloudCredentials) (*golangsdk.ServiceClient, error) {
	client, err := c.SFSTurboV1Client()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed create SFS Turbo V1 client: %s", err)
	}
	return client, nil
}
