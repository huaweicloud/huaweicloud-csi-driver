package common

import (
	"errors"
	"time"

	"github.com/chnsz/golangsdk"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	DefaultInitDelay = 2 * time.Second
	DefaultFactor    = 1.02
	DefaultSteps     = 30
)

func IsNotFound(err error) bool {
	return errors.As(err, &golangsdk.ErrDefault404{})
}

func WaitForCompleted(condition wait.ConditionFunc) error {
	backoff := wait.Backoff{
		Duration: DefaultInitDelay,
		Factor:   DefaultFactor,
		Steps:    DefaultSteps,
	}
	return wait.ExponentialBackoff(backoff, condition)
}
