package sfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	bytesInGiB = 1024 * 1024 * 1024
)

// TODO after feat-sfs-idempotent merged. Delete this method
func validateCreateVolumeRequest(req *csi.CreateVolumeRequest) error {
	if req.GetName() == "" {
		return errors.New("volume name cannot be empty")
	}

	reqCaps := req.GetVolumeCapabilities()
	if reqCaps == nil {
		return errors.New("volume capabilities cannot be empty")
	}

	/*
		for _, cap := range reqCaps {
			if cap.GetBlock() != nil {
				return errors.New("block access type not allowed")
			}
		}
	*/

	return nil
}

// TODO after feat-sfs-idempotent merged. Delete this method
func bytesToGiB(sizeInBytes int64) int {
	sizeInGiB := int(sizeInBytes / bytesInGiB)

	if int64(sizeInGiB)*bytesInGiB < sizeInBytes {
		// Round up
		return sizeInGiB + 1
	}

	return sizeInGiB
}

func Mount(source, target, mountOptions string) error {
	cmd := fmt.Sprintf("mount -t nfs -o vers=3,timeo=600,%s %s %s", mountOptions, source, target)
	_, err := Run(cmd)
	if err != nil {
		return err
	}
	return nil
}

func Unmount(target string) error {
	cmd := fmt.Sprintf("umount %s", target)
	_, err := Run(cmd)
	if err != nil {
		return err
	}
	return nil
}

func isMounted(target string) bool {
	cmd := fmt.Sprintf("mount | grep %s | grep -v grep | wc -l", target)
	out, err := Run(cmd)
	if err != nil {
		return false
	}
	if strings.TrimSpace(out) == "0" {
		return false
	}
	return true
}

func Run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to run cmd: " + cmd + ", with out: " + string(out) + ", with error: " + err.Error())
	}
	return string(out), nil
}

func makeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
