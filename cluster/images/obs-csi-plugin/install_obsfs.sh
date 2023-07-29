#!/bin/bash

checkLinuxOS() {
  if grep -i -q "EulerOS" /etc/os-release; then
    # EulerOS
    return 5
  elif [ -f "/etc/redhat-release" ]; then
    # CentOS or RHEL or EulerOS
    if grep -i -q "CentOS" /etc/redhat-release; then
      return 1
    elif grep -i -q "EulerOS" /etc/redhat-release; then
      return 5
    fi
    return 1
  elif [ -f "/etc/issue" ]; then
    # Ubuntu or Debian or EulerOS
    if grep -i -q "ubuntu" /etc/issue; then
      return 2
    elif grep -i -q "debian" /etc/issue; then
      return 3
    elif grep -i -q "Euler" /etc/os-release; then
      return 5
    fi
  elif [ -f "/etc/fedora-release" ]; then
    # Fedora
    return 4
  else
    # unknow OS
    return 0
  fi
}

buildAndInstallObsfs() {
  echo "[INFO] Install dependencies for OBSFS"

  if which yum &>/dev/null; then
    echo "[INFO] yum is available, try installing obsfs with yum"
    yum install -y gcc libstdc++-devel gcc-c++ fuse fuse-devel curl-devel libxml2-devel mailcap git automake make
    yum install -y openssl-devel
  elif which apt-get &>/dev/null; then
    echo "[INFO] apt-get is available, try installing obsfs with apt-get"
    apt-get install -y build-essential git libfuse-dev libcurl4-openssl-dev libxml2-dev mime-support automake libtool
    apt-get install -y pkg-config libssl-dev
  else
    echo "[ERROR] unsupported systems, please install obsfs manually"
  fi

  echo "[INFO] Unzip the code files"
  tar -xvf /var/lib/csi/huaweicloud-obs-obsfs.tar.gz -C /var/lib/csi/
  cd /var/lib/csi/huaweicloud-obs-obsfs

  echo "[INFO] Compile OBSFS"
  bash build.sh
  echo "[INFO] Install OBSFS"
  bash install_obsfs.sh

  echo "[INFO] Check version and check if it works"
  obsfs --version
}

installObsfs(){
  echo "[INFO] Install obsfs from ${1}"
  fileName=${1}
  rm -rf /var/lib/csi/${fileName}
  tar -zxvf /var/lib/csi/$fileName.tar.gz -C /var/lib/csi/

  cd /var/lib/csi/${fileName}
  bash install_obsfs.sh

  echo "[INFO] Check version and check if it works"
  obsfs --version

  n=`obsfs --version | grep "Storage Service File System" | wc -l`
  if [ "$n" = "0" ]; then
    echo "[WARN] Package doesn't work, try using compile and install"
    buildAndInstallObsfs
  fi
}

checkLinuxOS
osCode=$?
echo "[INFO] OS Code $osCode"

if [ "$osCode" = 1 ]; then
  echo "[INFO] OS is CentOS or EulerOS, use the package to install obsfs"
  yum install -y openssl-devel fuse fuse-devel
  installObsfs obsfs_CentOS7.6_amd64
elif [ "$osCode" = 2 ]; then
  echo "[INFO] OS is Ubuntu, use the package to install obsfs"
  apt-get install -y libfuse-dev libcurl4-openssl-dev
  installObsfs obsfs_Ubuntu16.04_amd64
elif [ "$osCode" = 5 ]; then
  buildAndInstallObsfs
else
  buildAndInstallObsfs
fi
