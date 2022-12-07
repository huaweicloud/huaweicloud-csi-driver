#!/usr/bin/env bash

set -o nounset
set -o pipefail

function usage() {
  echo "This script remove a k8s cluster"
  echo "      Usage: hack/remove-cluster.sh"
  echo "    Example: hack/remove-cluster.sh"
  echo
}

echo "Remove existing kubernetes"
kubeadm reset -f
echo "==========================================================================================="
yum list installed | grep kube
echo -e "===========================================================================================\n"
yum -y remove kubelet kubeadm kubectl kubernetes* cri-tools
