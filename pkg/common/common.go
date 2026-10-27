package common

import (
	"errors"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/bss/v2/orders"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

const (
	DefaultInitDelay = 2 * time.Second
	DefaultFactor    = 1.02
	DefaultSteps     = 30

	GbByteSize = 1024 * 1024 * 1024

	UnsubscribeResourcesAndPeriods = 1
)

func IsNotFound(err error) bool {
	return errors.As(err, &golangsdk.ErrDefault404{}) || status.Code(err) == codes.NotFound
}

func WaitForCompleted(condition wait.ConditionFunc) error {
	backoff := wait.Backoff{
		Duration: DefaultInitDelay,
		Factor:   DefaultFactor,
		Steps:    DefaultSteps,
	}
	return wait.ExponentialBackoff(backoff, condition)
}

func GetAZFromTopology(req *csi.TopologyRequirement, topologyKey string) string {
	for _, topology := range req.GetPreferred() {
		zone, exists := topology.GetSegments()[topologyKey]
		if exists {
			return zone
		}
	}

	for _, topology := range req.GetRequisite() {
		zone, exists := topology.GetSegments()[topologyKey]
		if exists {
			return zone
		}
	}
	return ""
}

func Unsubscribe(cc *config.CloudCredentials, resourceIds []string) error {
	client, err := cc.BssClient()
	if err != nil {
		return status.Errorf(codes.Internal, "Failed create BSS client: %s", err)
	}

	rst := orders.Unsubscribe(client, orders.UnsubscribeOpts{
		ResourceIds:     resourceIds,
		UnsubscribeType: UnsubscribeResourcesAndPeriods,
	})
	if rst.Err != nil {
		return status.Errorf(codes.Internal, "Failed to unsubscribe resources: %#v, error: %s", resourceIds, err)
	}

	return nil
}
