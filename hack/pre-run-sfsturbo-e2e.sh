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
  echo "This script build image and deploy SFS Turbo CSI on the kube cluster by kind."
  echo "      Usage: hack/pre-run-sfsturbo-e2e.sh <CLUSTER_NAME>"
  echo "    Example: hack/pre-run-sfsturbo-e2e.sh k8s-cluster"
  echo
}

if [[ -z "${HC_REGION}" || -z "${HC_ACCESS_KEY}" || -z "${HC_SECRET_KEY}" ]]; then
  echo "Error, please configure the HC_REGION, HC_ACCESS_KEY and HC_SECRET_KEY environment variables"
  usage
  exit 1
fi

set -o nounset

CLUSTER_NAME=$1
if [[ -z "${CLUSTER_NAME}" ]]; then
  echo "Error, CLUSTER_NAME can be empty"
  usage
  exit 1
fi

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
export VERSION=t`echo $RANDOM`

kubectl delete deployment/csi-sfsturbo-controller -n kube-system --ignore-not-found=true --context "${CLUSTER_NAME}"
kubectl wait --for=delete deployment/csi-sfsturbo-controller --timeout=60s --context "${CLUSTER_NAME}"

kubectl delete daemonset/csi-sfsturbo-plugin -n kube-system --ignore-not-found=true --context "${CLUSTER_NAME}"
kubectl wait --for=delete daemonset/csi-sfsturbo-plugin --timeout=60s --context "${CLUSTER_NAME}"

echo -e "\n>> Build SFS Turbo CSI plugin image"
make image-sfsturbo-csi-plugin
kind load docker-image "${REGISTRY_SERVER}/k8s-csi/sfsturbo-csi-plugin:${VERSION}" --name="${CLUSTER_NAME}"

echo -e "\n>> Deploy SFS Turbo CSI Plugin"
TEMP_PATH=$(mktemp -d)
echo ${TEMP_PATH}

cp -rf ${REPO_ROOT}/hack/deploy/sfsturbo/ ${TEMP_PATH}
## create Secret
sed -i'' -e "s/{{region}}/${HC_REGION}/g" "${TEMP_PATH}"/sfsturbo/cloud-config
sed -i'' -e "s/{{ak}}/${HC_ACCESS_KEY}/g" "${TEMP_PATH}"/sfsturbo/cloud-config
sed -i'' -e "s/{{sk}}/${HC_SECRET_KEY}/g" "${TEMP_PATH}"/sfsturbo/cloud-config
sed -i'' -e "s/{{vpc_id}}/${HC_VPC_ID}/g" "${TEMP_PATH}"/sfsturbo/cloud-config
sed -i'' -e "s/{{subnet_id}}/${HC_SUBNET_ID}/g" "${TEMP_PATH}"/sfsturbo/cloud-config
sed -i'' -e "s/{{security_group_id}}/${HC_SECURITY_GROUP_ID}/g" "${TEMP_PATH}"/sfsturbo/cloud-config
kubectl delete secret -n kube-system cloud-config --ignore-not-found=true --context "${CLUSTER_NAME}"
kubectl create secret -n kube-system generic cloud-config --from-file=${TEMP_PATH}/sfsturbo/cloud-config --context "${CLUSTER_NAME}"
rm -rf ${TEMP_PATH}/sfsturbo/cloud-config

## deploy plugin
image_url=${REGISTRY_SERVER}\\/k8s-csi\\/sfsturbo-csi-plugin:${VERSION}
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/sfsturbo/csi-sfsturbo-controller.yaml
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/sfsturbo/csi-sfsturbo-node.yaml
kubectl apply -f ${TEMP_PATH}/sfsturbo/rbac-csi-sfsturbo-controller.yaml --context "${CLUSTER_NAME}"
kubectl apply -f ${TEMP_PATH}/sfsturbo/rbac-csi-sfsturbo-node.yaml --context "${CLUSTER_NAME}"
kubectl apply -f ${TEMP_PATH}/sfsturbo/rbac-csi-sfsturbo-secret.yaml --context "${CLUSTER_NAME}"
kubectl apply -f ${TEMP_PATH}/sfsturbo/csi-sfsturbo-driver.yaml --context "${CLUSTER_NAME}"
kubectl apply -f ${TEMP_PATH}/sfsturbo/csi-sfsturbo-controller.yaml --context "${CLUSTER_NAME}"
kubectl apply -f ${TEMP_PATH}/sfsturbo/csi-sfsturbo-node.yaml --context "${CLUSTER_NAME}"

kubectl rollout status deployment csi-sfsturbo-controller -n kube-system --timeout=2m --context "${CLUSTER_NAME}"
kubectl rollout status daemonset csi-sfsturbo-plugin -n kube-system --timeout=2m --context "${CLUSTER_NAME}"
