#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
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

set -euo pipefail

repo="./deploy"

echo "Installing Huawei Cloud SFS CSI driver..."
kubectl apply -f $repo/rbac-csi-sfs-controller.yaml
kubectl apply -f $repo/rbac-csi-sfs-node.yaml
kubectl apply -f $repo/rbac-csi-sfs-secret.yaml
kubectl apply -f $repo/csi-sfs-controller.yaml
kubectl apply -f $repo/csi-sfs-driver.yaml
kubectl apply -f $repo/csi-sfs-node.yaml

echo 'Huawei Cloud SFS CSI driver installed successfully.'
