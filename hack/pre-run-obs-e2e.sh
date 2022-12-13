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
  echo "      Usage: hack/pre-run-obs-e2e.sh"
  echo "    Example: hack/pre-run-obs-e2e.sh"
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
export VERSION=t`echo $RANDOM`

kubectl delete deployment/csi-obs-controller -n kube-system --ignore-not-found=true
#kubectl wait --for=delete deployment/csi-obs-controller --timeout=60s

kubectl delete daemonset/csi-obs-plugin -n kube-system --ignore-not-found=true
#kubectl wait --for=delete daemonset/csi-obs-plugin --timeout=60s

echo -e "\n>> Build OBS CSI plugin image"
make image-obs-csi-plugin
#kind load docker-image "${REGISTRY_SERVER}/k8s-csi/obs-csi-plugin:${VERSION}"

echo -e "\n>> Deploy OBS CSI Plugin"
TEMP_PATH=$(mktemp -d)
echo ${TEMP_PATH}

cp -rf ${REPO_ROOT}/hack/deploy/obs/ ${TEMP_PATH}
## create Secret
sed -i'' -e "s/{{region}}/${HC_REGION}/g" "${TEMP_PATH}"/obs/cloud-config
sed -i'' -e "s/{{ak}}/${HC_ACCESS_KEY}/g" "${TEMP_PATH}"/obs/cloud-config
sed -i'' -e "s/{{sk}}/${HC_SECRET_KEY}/g" "${TEMP_PATH}"/obs/cloud-config

kubectl delete secret -n kube-system cloud-config --ignore-not-found=true
kubectl create secret -n kube-system generic cloud-config --from-file=${TEMP_PATH}/obs/cloud-config
rm -rf ${TEMP_PATH}/obs/cloud-config

## deploy plugin
image_url=${REGISTRY_SERVER}\\/k8s-csi\\/obs-csi-plugin:${VERSION}
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/obs/csi-obs-controller.yaml
sed -i'' -e "s/{{image_url}}/'${image_url}'/g" "${TEMP_PATH}"/obs/csi-obs-node.yaml
kubectl apply -f ${TEMP_PATH}/obs/rbac-csi-obs-controller.yaml
kubectl apply -f ${TEMP_PATH}/obs/rbac-csi-obs-node.yaml
kubectl apply -f ${TEMP_PATH}/obs/rbac-csi-obs-secret.yaml
kubectl apply -f ${TEMP_PATH}/obs/csi-obs-driver.yaml
kubectl apply -f ${TEMP_PATH}/obs/csi-obs-controller.yaml
kubectl apply -f ${TEMP_PATH}/obs/csi-obs-node.yaml

kubectl rollout status deployment csi-obs-controller -n kube-system --timeout=2m
kubectl rollout status daemonset csi-obs-plugin -n kube-system --timeout=2m

