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

export REGISTRY_SERVER=swr.ap-southeast-1.myhuaweicloud.com
export VERSION=`git rev-parse --short HEAD`
VERSION=${VERSION:-t`echo $RANDOM`}

kubectl delete deployment/csi-sfs-controller -n kube-system --ignore-not-found=true
kubectl wait --for=delete deployment/csi-sfs-controller --timeout=60s

kubectl delete daemonset/csi-sfs-node -n kube-system --ignore-not-found=true
kubectl wait --for=delete daemonset/csi-sfs-node --timeout=60s

echo -e "\n>> Build SFS CSI plugin image"
make image-sfs-csi-plugin

echo -e "\n>> Check cloud-config secret"
count=`kubectl get -n kube-system secret cloud-config | grep cloud-config | wc -l`
if [[ "$count" -ne 1 ]]; then
  echo "Please create a secret with the name: cloud-config."
  exit 1
fi

d="$"

echo -e "\n>> Deploy SFS CSI Plugin"
cat << EOF | kubectl apply -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: csi-sfs-controller-sa
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: sfs-external-provisioner-role
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["csinodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: sfs-csi-provisioner-binding
subjects:
  - kind: ServiceAccount
    name: csi-sfs-controller-sa
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: sfs-external-provisioner-role
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: sfs-external-attacher-role
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["csi.storage.k8s.io"]
    resources: ["csinodeinfos"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments/status"]
    verbs: ["patch"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: sfs-csi-attacher-binding
subjects:
  - kind: ServiceAccount
    name: csi-sfs-controller-sa
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: sfs-external-attacher-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: csi-sfs-node-sa
  namespace: kube-system
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-sfs-secret-role
  namespace: kube-system
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "create"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-sfs-secret-binding
  namespace: kube-system
subjects:
  - kind: ServiceAccount
    name: csi-sfs-controller-sa
    namespace: kube-system
  - kind: ServiceAccount
    name: csi-sfs-node-sa
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: csi-sfs-secret-role
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-sfs-secret-binding
  namespace: kube-system
subjects:
  - kind: ServiceAccount
    name: csi-sfs-controller-sa
    namespace: kube-system
  - kind: ServiceAccount
    name: csi-sfs-node-sa
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: csi-sfs-secret-role
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: sfs.csi.huaweicloud.com
spec:
  attachRequired: true
  podInfoOnMount: true
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: csi-sfs-controller
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: csi-sfs-controller
  template:
    metadata:
      labels:
        app: csi-sfs-controller
    spec:
      serviceAccountName: csi-sfs-controller-sa
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-cluster-critical
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Equal"
          value: "true"
          effect: "NoSchedule"
      containers:
        - name: csi-provisioner
          image: k8s.gcr.io/sig-storage/csi-provisioner:v3.1.0
          args:
            - "-v=5"
            - "--csi-address=$d(ADDRESS)"
            - "--timeout=3m"
            - "--default-fstype=ext4"
            - "--extra-create-metadata"
            - "--leader-election=true"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
        - name: csi-attacher
          image: k8s.gcr.io/sig-storage/csi-attacher:v3.3.0
          args:
            - "-v=5"
            - "--csi-address=$d(ADDRESS)"
            - "--timeout=3m"
            - "--leader-election=true"
          env:
            - name: ADDRESS
              value: /csi/csi.sock
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
        - name: sfs-csi-plugin
          imagePullPolicy: Never
          image: ${REGISTRY_SERVER}/k8s-csi/sfs-csi-plugin:${VERSION}
          args:
            - "--v=8"
            - "--logtostderr"
            - "--endpoint=$d(CSI_ENDPOINT)"
            - "--nodeid=$d(NODE_ID)"
            - "--cloud-config=$d(CLOUD_CONFIG)"
          ports:
            - containerPort: 28888
              name: healthz
              protocol: TCP
          livenessProbe:
            failureThreshold: 5
            httpGet:
              path: /healthz
              port: healthz
            initialDelaySeconds: 30
            timeoutSeconds: 10
            periodSeconds: 30
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: unix://csi/csi.sock
            - name: CLOUD_CONFIG
              value: /etc/sfs/cloud-config
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
            - mountPath: /etc/sfs/
              name: sfs-config
        - name: liveness-probe
          imagePullPolicy: Always
          image: k8s.gcr.io/sig-storage/livenessprobe:v2.6.0
          args:
            - --csi-address=/csi/csi.sock
            - --probe-timeout=3s
            - --health-port=28888
            - --v=5
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
      volumes:
        - name: socket-dir
          emptyDir: {}
        - name: sfs-config
          secret:
            secretName: cloud-config
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: csi-sfs-node
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: csi-sfs-node
  template:
    metadata:
      labels:
        app: csi-sfs-node
    spec:
      serviceAccountName: csi-sfs-node-sa
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-node-critical
      tolerations:
        - operator: "Exists"
      containers:
        - name: liveness-probe
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
          image: k8s.gcr.io/sig-storage/livenessprobe:v2.6.0
          args:
            - --csi-address=/csi/csi.sock
            - --probe-timeout=3s
            - --health-port=28889
            - --v=5
        - name: node-driver-registrar
          image: k8s.gcr.io/sig-storage/csi-node-driver-registrar:v2.5.0
          args:
            - --csi-address=$d(ADDRESS)
            - --kubelet-registration-path=$d(DRIVER_REG_SOCK_PATH)
            - --v=5
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "rm -rf /registration/sfs.csi.huaweicloud.com-reg.sock /csi/csi.sock"]
          env:
            - name: ADDRESS
              value: /csi/csi.sock
            - name: DRIVER_REG_SOCK_PATH
              value: /var/lib/kubelet/plugins/sfs.csi.huaweicloud.com/csi.sock
          volumeMounts:
            - name: socket-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
        - name: sfs
          imagePullPolicy: Never
          image: ${REGISTRY_SERVER}/k8s-csi/sfs-csi-plugin:${VERSION}
          args:
            - "--v=8"
            - "--logtostderr"
            - "--endpoint=$d(CSI_ENDPOINT)"
            - "--nodeid=$d(NODE_ID)"
            - "--cloud-config=$d(CLOUD_CONFIG)"
          ports:
            - containerPort: 28889
              name: healthz
              protocol: TCP
          livenessProbe:
            failureThreshold: 5
            httpGet:
              path: /healthz
              port: healthz
            initialDelaySeconds: 30
            timeoutSeconds: 10
            periodSeconds: 30
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: unix://csi/csi.sock
            - name: CLOUD_CONFIG
              value: /etc/sfs/cloud-config
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /csi
              name: socket-dir
            - mountPath: /var/lib/kubelet/
              mountPropagation: Bidirectional
              name: mountpoint-dir
            - mountPath: /etc/sfs/
              name: sfs-config
      volumes:
        - hostPath:
            path: /var/lib/kubelet/plugins/sfs.csi.huaweicloud.com
            type: DirectoryOrCreate
          name: socket-dir
        - hostPath:
            path: /var/lib/kubelet/
            type: DirectoryOrCreate
          name: mountpoint-dir
        - hostPath:
            path: /var/lib/kubelet/plugins_registry/
            type: DirectoryOrCreate
          name: registration-dir
        - secret:
            secretName: cloud-config
          name: sfs-config
EOF

kubectl rollout status deployment csi-sfs-controller -n kube-system --timeout=30s
kubectl rollout status daemonset csi-sfs-node -n kube-system --timeout=30s
