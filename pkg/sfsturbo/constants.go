package sfsturbo

// PVC annotation keys for the SFS Turbo sub-directory feature.
const (
	// VolumeAsAnnotation is the PVC annotation that enables the sub-directory
	// mode. When set to "absolute-path", the driver uses an existing SFS Turbo
	// instance and creates a sub-directory instead of provisioning a new share.
	VolumeAsAnnotation = "huaweicloud.com/volume-as"

	// VolumeAsAbsolutePath is the only valid value for VolumeAsAnnotation.
	VolumeAsAbsolutePath = "absolute-path"

	// SfsTurboShareIDAnnotation specifies the ID of the existing SFS Turbo
	// instance to use.
	SfsTurboShareIDAnnotation = "huaweicloud.com/sfsturbo-share-id"

	// PathAnnotation specifies the sub-directory path to create on the SFS
	// Turbo instance. Must be an absolute path (starts with "/").
	PathAnnotation = "huaweicloud.com/path"

	// ReclaimPolicyAnnotation controls what happens to the sub-directory when
	// the PVC is deleted. Valid values are "retain-volume-only" and "delete".
	ReclaimPolicyAnnotation = "huaweicloud.com/reclaim-policy"

	// ReclaimPolicyRetainVolumeOnly means: delete PVC → PV is deleted, but the
	// sub-directory and its quota are retained.
	ReclaimPolicyRetainVolumeOnly = "retain-volume-only"

	// ReclaimPolicyDelete means: delete PVC → PV, sub-directory, and its quota
	// are all deleted.
	ReclaimPolicyDelete = "delete"

	// EnableDirQuotaAnnotation controls whether directory quota is enabled.
	// Optional. When set to "true", the driver sets a capacity quota on the
	// sub-directory matching the PVC requested storage.
	EnableDirQuotaAnnotation = "huaweicloud.com/csi.enable-sfsturbo-dir-quota"
)

// CSI standard parameter keys injected by csi-provisioner with
// --extra-create-metadata. These are passed in CreateVolumeRequest.Parameters.
const (
	// PvcNameKey is the PVC name injected by csi-provisioner.
	PvcNameKey = "csi.storage.k8s.io/pvc/name"
	// PvcNamespaceKey is the PVC namespace injected by csi-provisioner.
	PvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
	// PvNameKey is the PV name injected by csi-provisioner.
	PvNameKey = "csi.storage.k8s.io/pv/name"
)

// VolumeContext keys passed from the controller to the node via the PV
// volumeAttributes.
const (
	// VolumeContextShareID stores the SFS Turbo share ID.
	VolumeContextShareID = "shareID"

	// VolumeContextPath stores the sub-directory path.
	VolumeContextPath = "path"

	// VolumeContextReclaimPolicy stores the reclaim policy.
	VolumeContextReclaimPolicy = "reclaimPolicy"

	// VolumeContextQuotaEnabled stores whether quota is enabled.
	VolumeContextQuotaEnabled = "quotaEnabled"

	// VolumeContextIsSubDir marks this volume as using the sub-directory mode.
	VolumeContextIsSubDir = "isSubDir"
)

// VolumeId separator for the sub-directory mode.
// Format: "<shareID>#<path>"
// When the VolumeId does not contain this separator, it is a legacy VolumeId
// (plain share UUID) and the original logic applies.
const volumeIDSep = "#"
