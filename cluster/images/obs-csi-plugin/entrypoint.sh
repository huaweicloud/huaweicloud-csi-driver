#!/bin/sh

HOST_CMD="/nsenter --mount=/proc/1/ns/mnt"
mkdir -p /var/lib/csi/

cp -f /obs-csi/install_s3fs.sh /var/lib/csi/install_s3fs.sh

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
