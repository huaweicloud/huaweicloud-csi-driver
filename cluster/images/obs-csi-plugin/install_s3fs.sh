#!/bin/bash

packageCmd=("yum" "dnf" "apt")

checkS3fsInstalled() {
  echo "[INFO] Check version and check if it works"
  isInstalled=1
  command -v s3fs >/dev/null 2>&1 || { isInstalled=0; }
  if [ ${isInstalled} = 0 ]; then
  	echo "[WARN] has no command named s3fs"
  	return ${isInstalled}
  fi
  echo "[INFO] s3fs has been installed"
  return ${isInstalled}
}

checkLinuxOS() {
  if grep -i -q "EulerOS" /etc/os-release; then
    # EulerOS
    return 5
  elif [ -f "/etc/redhat-release" ]; then
    # CentOS or RHEL or EulerOS
    if grep -i -q "CentOS" /etc/redhat-release; then
      return 1
    elif grep -i -q "Fedora" /etc/redhat-release; then
      return 4
    elif grep -i -q "EulerOS" /etc/redhat-release; then
      return 5
    fi
  elif [ -f "/etc/issue" ]; then
    # Ubuntu or Debian or Euler
    if grep -i -q "ubuntu" /etc/issue; then
      return 2
    elif grep -i -q "debian" /etc/issue; then
      return 3
    elif grep -i -q "Euler" /etc/os-release; then
      return 5
    fi
  elif [ -f "/etc/fedora-release" ]; then
    # Fedora
    if grep -i -q "Fedora" /etc/fedora-release; then
      return 4
    fi
  else
    # unknown OS
    return 0
  fi
}

installS3fs() {
  echo "[INFO] Install dependencies and s3fs"
  pgkManageCmd=${1}

  if which yum >/dev/null 2>&1 && [ ${pgkManageCmd} = "yum" ]; then
    echo "[INFO] command 'yum' is available, trying to install s3fs with yum"
    sudo yum -y install epel-release
    sudo yum -y install s3fs-fuse
    return
  elif which apt >/dev/null 2>&1 && [ ${pgkManageCmd} = "apt" ]; then
    echo "[INFO] command 'apt' is available, trying to install s3fs with apt"
    sudo apt -y install s3fs
    return
  elif which dnf >/dev/null 2>&1 && [ ${pgkManageCmd} = "dnf" ]; then
    echo "[INFO] command 'dnf' is available, trying to install s3fs with dnf"
    sudo dnf -y install s3fs-fuse
    return
  else
    echo "[ERROR] unsupported systems, please install s3fs manually"
  fi
}

# pre install
checkS3fsInstalled
isInstalled=$?
if [ ${isInstalled} = 1 ]; then
  exit 1
fi

# install
checkLinuxOS
osCode=$?
echo "[INFO] OS Code is ${osCode}"

if [ ${osCode} = 1 ] || [ ${osCode} = 5 ]; then
  echo "[INFO] OS is CentOS or EulerOS, use yum to install s3fs"
  installS3fs "yum"
elif [ ${osCode} = 2 ] || [ ${osCode} = 3 ]; then
  echo "[INFO] OS is Ubuntu or Debian, use apt to install s3fs"
  installS3fs "apt"
elif [ ${osCode} = 4 ]; then
  echo "[INFO] OS is Fedora, use dnf to install s3fs"
  installS3fs "dnf"
else
  echo "[WARN] Unknown OS, try to force installation of s3fs"
  for cmd in ${packageCmd[@]}; do
    installS3fs ${cmd}
    checkS3fsInstalled
    isInstalled=$?
    if [ ${isInstalled} = 1 ]; then
      exit 1
    fi
  done
fi

# post install
checkS3fsInstalled
isInstalled=$?
if [ ${isInstalled} = 1 ]; then
  exit 1
fi
echo "[ERROR] failed to install s3fs, please install s3fs manually"
