#!/bin/sh

centosCmd="cat /etc/redhat-release | grep -i centos | wc -l"
ubuntuCmd="cat /etc/issue | grep -i ubuntu | wc -l"
checkLinuxOS() {
  if [ -f "/etc/redhat-release" ]; then
    echo "/etc/redhat-release exists"
    if [ "$centosCmd" = 0 ]; then
      return 0
    else
      return 1
    fi
  fi

  if [ -f "/etc/issue" ]; then
    echo "/etc/issue exists"
    if [ "$ubuntuCmd" = 0 ]; then
      return 0
    else
      return 2
    fi
  fi
}
checkLinuxOS
osCode=$?
echo "OS Code $osCode"

HOST_CMD="/nsenter --mount=/proc/1/ns/mnt"
if [ "$osCode" = 1 ]; then
  echo "operation system is centos..."
  fileName=obsfs_CentOS7.6_amd64
  $HOST_CMD yum install -y openssl-devel fuse fuse-devel
elif [ "$osCode" = 2 ]; then
  echo "operation system is ubuntu..."
  fileName=obsfs_Ubuntu16.04_amd64
  $HOST_CMD apt-get install -y libfuse-dev libcurl4-openssl-dev
else
  echo "operation system not support..."
  exit
fi

echo "Starting deploy obs csi-plugin...."
mkdir -p /dev/csi-tool/
tar -zxvf /root/$fileName.tar.gz -C /dev/csi-tool/
$HOST_CMD cp -r /dev/csi-tool/$fileName/. ./
$HOST_CMD bash install_obsfs.sh

echo "Starting run obs socket server...."
cp /bin/socket-server /dev/csi-tool/socket-server
chmod 755 /dev/csi-tool/socket-server
$HOST_CMD /dev/csi-tool/socket-server >log.out 2>&1 &

echo "Starting run obs csi plugin...."
/bin/obs-csi-plugin $@
