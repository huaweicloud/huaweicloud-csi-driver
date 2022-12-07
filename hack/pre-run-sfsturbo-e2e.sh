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

function usage() {
  echo "This script build image and deploy SFS Turbo CSI on the kube cluster by kind."
  echo "      Usage: hack/pre-run-sfsturbo-e2e.sh"
  echo "    Example: hack/pre-run-sfsturbo-e2e.sh"
  echo
}

HC_REGION=${HC_REGION:-""}
HC_ACCESS_KEY=${HC_ACCESS_KEY:-""}
HC_SECRET_KEY=${HC_SECRET_KEY:-""}
HC_VPC_ID=${HC_VPC_ID:-""}
HC_SUBNET_ID=${HC_SUBNET_ID:-""}
HC_SECURITY_GROUP_ID=${HC_SECURITY_GROUP_ID:-""}
if [[ -z "${HC_REGION}" || -z "${HC_ACCESS_KEY}" || -z "${HC_SECRET_KEY}" || -z "${HC_VPC_ID}" || -z "${HC_SUBNET_ID}" || -z "${HC_SECURITY_GROUP_ID}" ]]; then
  echo "Error, please configure the HC_REGION, HC_ACCESS_KEY and HC_SECRET_KEY environment variables"
  usage
  exit 1
fi

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
export VERSION=v$(echo $RANDOM | sha1sum |cut -c 1-5)

#kubectl delete deployment/csi-sfsturbo-controller -n kube-system --ignore-not-found=true
#kubectl wait --for=delete deployment/csi-sfsturbo-controller --timeout=60s

#kubectl delete daemonset/csi-sfsturbo-plugin -n kube-system --ignore-not-found=true
#kubectl wait --for=delete daemonset/csi-sfsturbo-plugin --timeout=60s

echo -e "\n>> Build SFS Turbo CSI plugin image"
make image-sfsturbo-csi-plugin

TEMP_PATH=$(mktemp -d)
echo ${TEMP_PATH}

is_containerd=`command -v containerd`
echo "is_containerd: ${is_containerd}"
if [[ -x ${is_containerd} ]]; then
  docker save -o "${TEMP_PATH}/sfsturbo-csi-plugin.tar" ${REGISTRY_SERVER}/k8s-csi/sfsturbo-csi-plugin:${VERSION}
  ctr -n=k8s.io i import ${TEMP_PATH}/sfsturbo-csi-plugin.tar
  rm -rf ${TEMP_PATH}/sfsturbo-csi-plugin.tar
fi
#kind load docker-image "${REGISTRY_SERVER}/k8s-csi/sfsturbo-csi-plugin:${VERSION}" --name="${CLUSTER_NAME}"

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
