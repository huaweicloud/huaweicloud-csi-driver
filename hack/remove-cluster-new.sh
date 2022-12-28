#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function usage() {
  echo "This script remove a k8s cluster"
  echo "      Usage: hack/remove-cluster.sh"
  echo "    Example: hack/remove-cluster.sh"
  echo
}

echo "Remove existing kubernetes"
for service in kube-apiserver kube-controller-manager kubectl kubelet kube-proxy kube-scheduler;
do
  systemctl stop $service
done

kubeadm reset -f
rm -rf ~/.kube/
rm -rf /etc/kubernetes/
rm -rf /etc/systemd/system/kubelet.service.d
rm -rf /etc/systemd/system/kubelet.service
rm -rf /usr/bin/kube*
rm -rf /etc/cni
rm -rf /opt/cni
rm -rf /var/lib/etcd
rm -rf /var/etcd
yum clean all
yum -y remove kube* > /dev/null 2>&1
