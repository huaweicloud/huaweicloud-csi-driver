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
set -o pipefail
set -o nounset

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

kubectl delete -f ${REPO_ROOT}/deploy/evs-csi-plugin/kubernetes/csi-evs-controller.yaml
kubectl delete -f ${REPO_ROOT}/deploy/evs-csi-plugin/kubernetes/csi-evs-driver.yaml
kubectl delete -f ${REPO_ROOT}/deploy/evs-csi-plugin/kubernetes/csi-evs-node.yaml
kubectl delete -f ${REPO_ROOT}/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-controller.yaml --ignore-not-found=true
kubectl delete -f ${REPO_ROOT}/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-node.yaml --ignore-not-found=true
kubectl delete -f ${REPO_ROOT}/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-secret.yaml --ignore-not-found=true
