package test

import (
	"testing"

	"github.com/chnsz/golangsdk"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/common"
)

func TestIsNotFound(t *testing.T) {
	err404 := golangsdk.ErrDefault404{}
	if !common.IsNotFound(err404) {
		t.Errorf("Error in TestIsNotFound")
	}
}

func TestWaitForCompleted(t *testing.T) {
	n := 0

	err := common.WaitForCompleted(func() (bool, error) {
		if n++; n == 2 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Errorf("Error in TestWaitForCompleted")
	}
}
