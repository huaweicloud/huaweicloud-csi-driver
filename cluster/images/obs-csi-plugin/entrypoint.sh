#!/bin/sh

HOST_CMD="/nsenter --mount=/proc/1/ns/mnt"
checkLinuxOS() {
  osId=$($HOST_CMD cat /etc/redhat-release | grep -i -c centos)
  if [ "$osId" != 0 ]; then
    return 1
  fi

  osId=$($HOST_CMD cat /etc/issue | grep -i -c ubuntu)
  if [ "$osId" != 0 ]; then
      return 2
  fi
  return 0
}
checkLinuxOS
osCode=$?
echo "OS Code $osCode"

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

echo "Starting install obsfs...."
mkdir -p /dev/csi-tool/
tar -zxvf /root/$fileName.tar.gz -C /dev/csi-tool/
$HOST_CMD cp -r /dev/csi-tool/$fileName/. ./
$HOST_CMD bash install_obsfs.sh

echo "Starting install obs socket-server...."
rm -rf /dev/csi-tool/socket-server
rm -rf /dev/csi-tool/connector.sock
cp /bin/socket-server /dev/csi-tool/socket-server
chmod 755 /dev/csi-tool/socket-server

echo "Starting install obs socket service...."
cp /bin/socket-server.service /dev/csi-tool/socket-server.service
$HOST_CMD cp /dev/csi-tool/socket-server.service /lib/systemd/system/socket-server.service
$HOST_CMD systemctl daemon-reload
$HOST_CMD systemctl enable socket-server.service
$HOST_CMD systemctl restart socket-server.service

echo "Starting run obs csi plugin...."
/bin/obs-csi-plugin $@
