package evs

const (
	// PvcNameTag in annotations
	PvcNameTag = "csi.storage.k8s.io/pvc/name"
	// PvcNsTag in annotations
	PvcNsTag = "csi.storage.k8s.io/pvc/namespace"
	// PvNameKey key
	PvNameKey = "csi.storage.k8s.io/pv/name"

	// CsiClusterNodeIDKey in volume metadata
	CsiClusterNodeIDKey = "evs.csi.huaweicloud.com/nodeId"
	// CsiClusterNodeIDKey in volume metadata
	DssIDKey = "dedicated_storage_id"
	// CreateForVolumeIDKey in volume metadata
	CreateForVolumeIDKey = "create_for_volume_id"
	// HwPassthroughKey in volume metadata
	HwPassthroughKey = "hw:passthrough"
)
