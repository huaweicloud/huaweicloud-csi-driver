#!/usr/bin/env bash

set -o errexit
set -o pipefail

function usage() {
  echo "This script starts a kube cluster by kubelet."
  echo "      Usage: hack/switch-cluster.sh <CLUSTER_VERSION>"
  echo "    Example: hack/switch-cluster.sh 1.20"
  echo
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

#kubectl config use-context "k8s-${1}"
kubectl version
kubectl version --short
