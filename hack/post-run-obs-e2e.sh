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

#export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

echo -e "\n>> Deploy OBS CSI Plugin"
TEMP_PATH=$(mktemp -d)
echo ${TEMP_PATH}

cp -rf ${REPO_ROOT}/hack/deploy/obs/ ${TEMP_PATH}
## deploy plugin
image_url=${REGISTRY_SERVER}\\/k8s-csi\\/obs-csi-plugin:${VERSION}
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/obs/csi-obs-controller.yaml
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/obs/csi-obs-node.yaml

kubectl delete -f ${TEMP_PATH}/obs/csi-obs-controller.yaml
kubectl delete -f ${TEMP_PATH}/obs/csi-obs-driver.yaml
kubectl delete -f ${TEMP_PATH}/obs/csi-obs-node.yaml
kubectl delete -f ${TEMP_PATH}/obs/rbac-csi-obs-controller.yaml
kubectl delete -f ${TEMP_PATH}/obs/rbac-csi-obs-node.yaml
kubectl delete secret -n kube-system cloud-config --ignore-not-found=true

