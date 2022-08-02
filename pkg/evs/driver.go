package evs

import (
	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

const (
	driverName  = "evs.csi.huaweicloud.com"
)

var (
	// CSI spec version
	specVersion = "1.3.0"
	// Driver version
	Version = "1.0.0"
)

type EvsDriver struct {
	name       string
	nodeID     string
	version    string
	endpoint   string
	cluster    string
	shareProto string
	cloud      config.CloudCredentials

	ids *identityServer
	cs  *ControllerServer
	ns  *nodeServer

	vcap  []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
	nscap []*csi.NodeServiceCapability
}

func NewDriver(cloud *config.CloudCredentials, endpoint, cluster, nodeID string) *EvsDriver {
	return nil
}
