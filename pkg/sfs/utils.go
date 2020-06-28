package sfs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

const (
	bytesInGiB = 1024 * 1024 * 1024
)

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func NewNodeServiceCapability(cap csi.NodeServiceCapability_RPC_Type) *csi.NodeServiceCapability {
	return &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func NewVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

func NewControllerServer(d *SfsDriver) *controllerServer {
	return &controllerServer{
		Driver: d,
	}
}

func NewIdentityServer(d *SfsDriver) *identityServer {
	return &identityServer{
		Driver: d,
	}
}

func NewNodeServer(d *SfsDriver) *nodeServer {
	return &nodeServer{
		Driver:   d,
	}
}

func RunControllerandNodePublishServer(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer) {

	s := NewNonBlockingGRPCServer()
	s.Start(endpoint, ids, cs, ns)
	s.Wait()
}

func ParseEndpoint(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
}

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	klog.V(3).Infof("GRPC call: %s", info.FullMethod)
	klog.V(5).Infof("GRPC request: %+v", req)
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	} else {
		klog.V(5).Infof("GRPC response: %+v", resp)
	}
	return resp, err
}

//
// Controller service request validation
//

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

// RoundUpSize calculates how many allocation units are needed to accommodate
// a volume of given size. E.g. when user wants 1500MiB volume, while AWS EBS
// allocates volumes in gibibyte-sized chunks,
// RoundUpSize(1500 * 1024*1024, 1024*1024*1024) returns '2'
// (2 GiB is the smallest allocatable volume that can hold 1500MiB)
func RoundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	roundedUp := volumeSizeBytes / allocationUnitBytes
	if volumeSizeBytes%allocationUnitBytes > 0 {
		roundedUp++
	}
	return roundedUp
}

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
