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

export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
export VERSION=t`echo $RANDOM`

kubectl delete deployment/csi-obs-controller -n kube-system --ignore-not-found=true
#kubectl wait --for=delete deployment/csi-obs-controller --timeout=60s

kubectl delete daemonset/csi-obs-plugin -n kube-system --ignore-not-found=true
#kubectl wait --for=delete daemonset/csi-obs-plugin --timeout=60s

echo -e "\n>> Build OBS CSI plugin image"
make image-obs-csi-plugin

TEMP_PATH=$(mktemp -d)
is_containerd=`command -v containerd`
echo "is_containerd: ${is_containerd}"
if [[ -x ${is_containerd} ]]; then
  docker save -o "${TEMP_PATH}/obs-csi-plugin.tar" ${REGISTRY_SERVER}/k8s-csi/obs-csi-plugin:${VERSION}
  ctr -n=k8s.io i import ${TEMP_PATH}/obs-csi-plugin.tar
  rm -rf ${TEMP_PATH}/obs-csi-plugin.tar
fi

echo -e "\n>> Deploy OBS CSI Plugin"
cp -rf ${REPO_ROOT}/hack/deploy/obs/ ${TEMP_PATH}

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
