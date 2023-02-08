#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function usage() {
  echo "This script create a kube cluster by kubeadm."
  echo "      Usage: hack/create-cluster.sh <CLUSTER_VERSION>"
  echo "    Example: hack/create-cluster.sh 1.21.10"
  echo
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

CLUSTER_VERSION=$1
if [[ -z "${CLUSTER_VERSION}" ]]; then
  usage
  exit 1
fi

/root/hack/create-cluster.sh ${CLUSTER_VERSION}
