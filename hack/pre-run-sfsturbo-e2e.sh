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

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
export VERSION=v$(echo $RANDOM | sha1sum |cut -c 1-5)

echo -e "\n>> Build SFS Turbo CSI plugin image"
make image-sfsturbo-csi-plugin

TEMP_PATH=$(mktemp -d)

is_containerd=`command -v containerd`
echo "is_containerd: ${is_containerd}"
if [[ -x ${is_containerd} ]]; then
  docker save -o "${TEMP_PATH}/sfsturbo-csi-plugin.tar" ${REGISTRY_SERVER}/k8s-csi/sfsturbo-csi-plugin:${VERSION}
  ctr -n=k8s.io i import ${TEMP_PATH}/sfsturbo-csi-plugin.tar
  rm -rf ${TEMP_PATH}/sfsturbo-csi-plugin.tar
fi

echo -e "\n>> Deploy SFS Turbo CSI Plugin"
cp -rf ${REPO_ROOT}/hack/deploy/sfsturbo/ ${TEMP_PATH}

## deploy plugin
image_url=${REGISTRY_SERVER}\\/k8s-csi\\/sfsturbo-csi-plugin:${VERSION}
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/sfsturbo/csi-sfsturbo-controller.yaml
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/sfsturbo/csi-sfsturbo-node.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/rbac-csi-sfsturbo-controller.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/rbac-csi-sfsturbo-node.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/rbac-csi-sfsturbo-secret.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/csi-sfsturbo-driver.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/csi-sfsturbo-controller.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/csi-sfsturbo-node.yaml

kubectl rollout status deployment csi-sfsturbo-controller -n kube-system --timeout=3m
kubectl rollout status daemonset csi-sfsturbo-plugin -n kube-system --timeout=3m
