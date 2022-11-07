#!/usr/bin/env bash

# Copyright 2022 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

echo "Run post E2E"
# delete sfs objects
kubectl delete serviceaccount csi-sfs-controller-sa csi-sfs-node-sa -n kube-system
kubectl delete clusterrole sfs-external-provisioner-role sfs-external-attacher-role csi-sfs-secret-role
kubectl delete clusterrolebinding sfs-csi-provisioner-binding sfs-csi-attacher-binding csi-sfs-secret-binding
kubectl delete csidriver sfs.csi.huaweicloud.com
kubectl delete deployment csi-sfs-controller -n kube-system
kubectl delete daemonset csi-sfs-node -n kube-system
