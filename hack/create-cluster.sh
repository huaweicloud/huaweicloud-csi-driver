#!/usr/bin/env bash

set -o errexit
set -o pipefail

function usage() {
  echo "This script starts a kube cluster by kubelet."
  echo "      Usage: hack/create-cluster.sh <CLUSTER_VERSION>"
  echo "    Example: hack/create-cluster.sh 1.20.16"
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

function install_k8s(){
  echo "Disable swap"
  tee /etc/sysctl.d/k8s.conf <<-'EOF'
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

  sysctl -p /etc/sysctl.d/k8s.conf

  echo "Set kubernetes repository"
  tee /etc/yum.repos.d/kubernetes.repo <<-'EOF'
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF

  yum clean all
  yum -y makecache

  echo "Install kubernetes ${CLUSTER_VERSION}"
  yum -y install kubelet-${CLUSTER_VERSION} kubeadm-${CLUSTER_VERSION} kubectl-${CLUSTER_VERSION}
  systemctl enable kubelet && systemctl start kubelet

  echo "export KUBECONFIG=/etc/kubernetes/admin.conf" >>~/.bash_profile
  source ~/.bash_profile
}

function init_master {
  echo "Initialize the master node"
  kubeadm reset -f
  rm -rf /etc/cni/net.d
  rm -rf $HOME/.kube/config
  rm -rf /etc/containerd/config.toml
  systemctl restart containerd
  sleep 3
  kubeadm init --service-cidr=10.1.0.0/16 --pod-network-cidr=10.244.0.0/16

  echo "Remove taint from master node"
  set +o errexit
  kubectl taint nodes --all node-role.kubernetes.io/master-
  kubectl taint nodes --all node-role.kubernetes.io/control-plane-
  kubectl get no -o yaml | grep taint -A 5

  echo "Install kube-flannel"
  kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
  echo "Wait for kube-flannel to be available"
  sleep 2
  kubectl rollout status daemonset kube-flannel-ds -n kube-flannel --timeout=3m
  kubectl get all -A
}

function check_cmd_exist {
  local CMD=$(command -v ${1})
  if [[ ! -x ${CMD} ]]; then
    echo "Please install \"${1}\" and verify they are in \$PATH."
    exit 1
  fi
}

check_cmd_exist docker
install_k8s
init_master
