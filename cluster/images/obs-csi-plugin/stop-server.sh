#!/bin/bash

HOST_CMD="/nsenter --mount=/proc/1/ns/mnt"

$HOST_CMD systemctl daemon-reload
$HOST_CMD systemctl disable csi-connector.service
$HOST_CMD systemctl stop csi-connector.service
$HOST_CMD rm -rf /etc/systemd/system/csi-connector.service
$HOST_CMD systemctl daemon-reload
