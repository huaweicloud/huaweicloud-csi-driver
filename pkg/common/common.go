package common

import (
	"errors"
	"time"

	"github.com/chnsz/golangsdk"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	DefaultInitDelay = 2 * time.Second
	DefaultFactor    = 1.02
	DefaultSteps     = 30

	GbByteSize = 1024 * 1024 * 1024
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
