#!/usr/bin/env bash
# Copyright 2022 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

echo -e "====== Start Test SFS(share) "

function before_test() {
  testRes="false"
  cat << EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: sfs-sc
provisioner: sfs.csi.huaweicloud.com
reclaimPolicy: Delete
EOF

  cat << EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pvc-sfs
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: sfs-sc
EOF
}

function post_test(){
  kubectl delete pod nginx-sfs --force --grace-period=0 --ignore-not-found=true

  kubectl delete pvc pvc-sfs
  kubectl delete sc sfs-sc
}

function do_test() {
  kubectl delete pod nginx-sfs --force --grace-period=0 --ignore-not-found=true
  cat << EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: nginx-sfs
  labels:
    app: sfs-test
spec:
  containers:
  - image: nginx
    name: nginx-sfs
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo $(date -u) >> /mnt/sfs/outfile; sleep 5; done"]
    volumeMounts:
    - mountPath: /mnt/sfs
      name: sfs-data
  volumes:
  - name: sfs-data
    persistentVolumeClaim:
      claimName: pvc-sfs
EOF

  kubectl wait --for=condition=Ready pod/nginx-sfs --timeout=60s
  sleep 20

  kubectl delete pod nginx-sfs
  kubectl wait --for=delete pod/nginx-sfs --timeout=60s

  echo -e "------ PASS: SFS Test\n"
}
echo ">> Start run SFS CSI test"
echo ">> Before test"
before_test
echo ">> Run test"
do_test
echo ">> After test"
post_test
