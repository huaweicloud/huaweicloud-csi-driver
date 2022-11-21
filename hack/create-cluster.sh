#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function usage() {
  echo "This script starts a kube cluster by kind."
  echo "      Usage: hack/create-cluster.sh <CLUSTER_NAME>"
  echo "    Example: hack/create-cluster.sh host"
  echo
}

CLUSTER_VERSION=${CLUSTER_VERSION:-"kindest/node:v1.21.10"}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

CLUSTER_NAME=$1
if [[ -z "${CLUSTER_NAME}" ]]; then
  usage
  exit 1
fi

# check kind
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

function rand() {
    min=$1
    max=$(($2-$min+1))
    num=$(date +%s%N)
    echo $(($num%$max+$min))
}

function generate_cidr() {
    local val=$(rand 1 254)
    echo "$1.${val}.0.0/16"
}

POD_CIDR=${POD_CIDR:-""}
if [[ -z "${POD_CIDR}" ]]; then
    POD_CIDR=$(generate_cidr 10)
fi

SERVICE_CIDR=${SERVICE_CIDR:-""}
if [[ -z "${SERVICE_CIDR}" ]]; then
    SERVICE_CIDR=$(generate_cidr 100)
fi

#generate for kindClusterConfig
TEMP_PATH=$(mktemp -d)
cat <<EOF > ${TEMP_PATH}/${CLUSTER_NAME}-kind-config.yaml
kind: Cluster
apiVersion: "kind.x-k8s.io/v1alpha4"
networking:
  podSubnet: ${POD_CIDR}
  serviceSubnet: ${SERVICE_CIDR}
nodes:
  - role: control-plane
  - role: worker
EOF

nohup ${REPO_ROOT}/hack/delete-cluster.sh "${CLUSTER_NAME}" >/dev/null 2>&1 &
sleep 10
kind create cluster --name "${CLUSTER_NAME}" --image="${CLUSTER_VERSION}" --config="${TEMP_PATH}"/"${CLUSTER_NAME}"-kind-config.yaml

# Kind cluster's context name contains a "kind-" prefix by default.
# Change context name to cluster name.
kubectl config rename-context "kind-${CLUSTER_NAME}" "${CLUSTER_NAME}"

container_ip=$(docker inspect --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${CLUSTER_NAME}-control-plane")
kubectl config set-cluster "kind-${CLUSTER_NAME}" --server="https://${container_ip}:6443"

echo "cluster \"${CLUSTER_NAME}\" is created successfully!"
