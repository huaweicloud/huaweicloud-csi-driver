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

function usage() {
  echo "This script build image and deploy EVS CSI on the kube cluster."
  echo "      Usage: hack/pre-run-evs-e2e.sh"
  echo "    Example: hack/pre-run-evs-e2e.sh"
  echo
}

if [[ -z "${HC_REGION}" || -z "${HC_ACCESS_KEY}" || -z "${HC_SECRET_KEY}" ]]; then
  echo "Error, please configure the HC_REGION, HC_ACCESS_KEY and HC_SECRET_KEY environment variables"
  usage
  exit 1
fi

set -o nounset

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
export VERSION=v$(echo $RANDOM | sha1sum |cut -c 1-5)

echo -e "\n>> Build EVS CSI plugin image"
make image-evs-csi-plugin

TEMP_PATH=$(mktemp -d)
echo ${TEMP_PATH}

is_containerd=`command -v containerd`
echo "is_containerd: ${is_containerd}"
if [[ -x ${is_containerd} ]]; then
  docker save -o "${TEMP_PATH}/evs-csi-plugin.tar" ${REGISTRY_SERVER}/k8s-csi/evs-csi-plugin:${VERSION}
  ctr -n=k8s.io i import ${TEMP_PATH}/evs-csi-plugin.tar
  rm -rf ${TEMP_PATH}/evs-csi-plugin.tar
fi

echo -e "\n>> Deploy EVS CSI Plugin"
cp -rf ${REPO_ROOT}/hack/deploy/evs/ ${TEMP_PATH}
image_url=${REGISTRY_SERVER}\\/k8s-csi\\/evs-csi-plugin:${VERSION}
## deploy plugin
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/evs/csi-evs-controller.yaml
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/evs/csi-evs-node.yaml
kubectl apply -f ${TEMP_PATH}/evs/rbac-csi-evs-controller.yaml
kubectl apply -f ${TEMP_PATH}/evs/rbac-csi-evs-node.yaml
kubectl apply -f ${TEMP_PATH}/evs/rbac-csi-evs-secret.yaml
kubectl apply -f ${TEMP_PATH}/evs/csi-evs-driver.yaml
kubectl apply -f ${TEMP_PATH}/evs/csi-evs-controller.yaml
kubectl apply -f ${TEMP_PATH}/evs/csi-evs-node.yaml

kubectl rollout status deployment csi-evs-controller -n kube-system --timeout=3m
kubectl rollout status daemonset csi-evs-plugin -n kube-system --timeout=3m
