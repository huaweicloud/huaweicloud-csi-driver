package evs

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
)

const (
	driverName  = "evs.csi.huaweicloud.com"
	topologyKey = "topology." + driverName + "/zone"
)

var (
	// specVersion CSI spec version
	specVersion = "1.3.0"
	// Version Driver version
	Version = "1.0.0"
)

type EvsDriver struct {
	name       string
	nodeID     string
	version    string
	endpoint   string
	cluster    string
	shareProto string

	cloudCredentials *config.CloudCredentials

	ids *identityServer
	cs  *ControllerServer
	ns  *nodeServer

	vcap  []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
	nscap []*csi.NodeServiceCapability
}

func (d *EvsDriver) GetControllerServer() *ControllerServer {
	return d.cs
}

func (d *EvsDriver) GetIdentityServer() *identityServer {
	return d.ids
}

func (d *EvsDriver) GetNodeServer() *nodeServer {
	return d.ns
}

func NewDriver(cc *config.CloudCredentials, endpoint, cluster, nodeID string) *EvsDriver {
	d := &EvsDriver{}
	d.name = driverName
	d.version = fmt.Sprintf("%s@%s", Version, specVersion)
	d.endpoint = endpoint
	d.cluster = cluster
	d.nodeID = nodeID
	d.cloudCredentials = cc

	log.Info("Driver: ", d.name)
	log.Info("Driver version: ", d.version)
	log.Info("CSI Spec version: ", specVersion)

	d.addControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
			csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
			csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
			//csi.ControllerServiceCapability_RPC_CLONE_VOLUME, // the API does not support clone feature
			csi.ControllerServiceCapability_RPC_LIST_VOLUMES_PUBLISHED_NODES,
			csi.ControllerServiceCapability_RPC_GET_VOLUME,
		})
	d.addVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	d.addNodeServiceCapabilities(
		[]csi.NodeServiceCapability_RPC_Type{
			csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
			csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		})

	d.ids = &identityServer{Driver: d}
	d.cs = &ControllerServer{Driver: d}
	d.ns = &nodeServer{Driver: d}

	return d
}

func (d *EvsDriver) addControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	csc := make([]*csi.ControllerServiceCapability, 0)
	for _, c := range cl {
		log.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: c,
				},
			},
		})
	}
	d.cscap = csc
}

func (d *EvsDriver) addVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability_AccessMode {
	vca := make([]*csi.VolumeCapability_AccessMode, 0)
	for _, c := range vc {
		log.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, &csi.VolumeCapability_AccessMode{Mode: c})
	}
	d.vcap = vca
	return vca
}

func (d *EvsDriver) addNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) {
	nsc := make([]*csi.NodeServiceCapability, 0)
	for _, n := range nl {
		log.Infof("Enabling node service capability: %v", n.String())
		nsc = append(nsc, &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: n,
				},
			},
		})
	}
	d.nscap = nsc
}

func (d *EvsDriver) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, csc := range d.cscap {
		if c == csc.GetRpc().GetType() {
			return nil
		}
	}
	return status.Error(codes.InvalidArgument, fmt.Sprintf("%s", c))
}

func (d *EvsDriver) GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	return d.vcap
}

func (d *EvsDriver) Run() {
	s := NewNonBlockingGRPCServer()
	s.Start(d.endpoint, d.ids, d.cs, d.ns)
	s.Wait()
}
