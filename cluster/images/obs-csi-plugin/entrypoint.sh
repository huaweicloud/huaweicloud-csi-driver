#!/bin/sh

HOST_CMD="/nsenter --mount=/proc/1/ns/mnt"
mkdir -p /var/lib/csi/

cp -f /obs-csi/huaweicloud-obs-obsfs.tar.gz /var/lib/csi/huaweicloud-obs-obsfs.tar.gz
cp -f /obs-csi/obsfs_CentOS7.6_amd64.tar.gz /var/lib/csi/obsfs_CentOS7.6_amd64.tar.gz
cp -f /obs-csi/obsfs_Ubuntu16.04_amd64.tar.gz /var/lib/csi/obsfs_Ubuntu16.04_amd64.tar.gz
cp -f /obs-csi/install_obsfs.sh /var/lib/csi/install_obsfs.sh

echo "Starting install obs csi-connector-server...."
$HOST_CMD systemctl stop csi-connector.service
rm -rf /var/lib/csi/connector.sock
cp -f /obs-csi/csi-connector-server /var/lib/csi/csi-connector-server
cp -f /obs-csi/csi-connector.service /var/lib/csi/csi-connector.service
chmod 755 /var/lib/csi/csi-connector-server

echo "Run csi-connector-server...."
$HOST_CMD cp -f /var/lib/csi/csi-connector.service /etc/systemd/system/csi-connector.service
$HOST_CMD systemctl daemon-reload
$HOST_CMD systemctl enable csi-connector.service
$HOST_CMD systemctl restart csi-connector.service
$HOST_CMD systemctl status csi-connector.service

echo "Starting run obs csi plugin...."
/obs-csi/obs-csi-plugin $@
