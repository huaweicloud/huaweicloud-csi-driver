/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package obs

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
