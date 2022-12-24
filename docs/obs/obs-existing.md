# Use Existing Bucket

Use an existing bucket, provide a custom PV to bind PVC. When removing PVC or PV, the bucket resource will remain.

## Prerequisites

- kubernetes, OBS CSI Driver

## How to use

### Step 1: Create PV

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/use-existing-bucket/pv.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/use-existing-bucket/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/use-existing-bucket/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME               READY   STATUS    RESTARTS   AGE
nginx-obs-static   1/1     Running   0          75s

```

```
# kubectl get pvc
NAME             STATUS   VOLUME          CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-obs-static   Bound    pv-obs-static   5Gi        RWX                           25s
```

```
# kubectl get pv
NAME            CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
pv-obs-static   5Gi        RWX            Delete           Bound    default/pvc-obs-static                           32s
```

```
# kubectl get pv pv-obs-static -o yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"PersistentVolume","metadata":{"annotations":{},"name":"pv-obs-static"},"spec":{"accessModes":["ReadWriteMany"],"capacity":{"storage":"5Gi"},"csi":{"driver":"obs.csi.huaweicloud.com","volumeHandle":"custom-bucket"},"persistentVolumeReclaimPolicy":"Delete"}}
    pv.kubernetes.io/bound-by-controller: "yes"
  creationTimestamp: "2022-12-23T06:48:19Z"
  finalizers:
  - kubernetes.io/pv-protection
  name: pv-obs-static
  resourceVersion: "819915"
  uid: 3a1fa3f1-c054-4d3b-8434-959290d0b6af
spec:
  accessModes:
  - ReadWriteMany
  capacity:
    storage: 5Gi
  claimRef:
    apiVersion: v1
    kind: PersistentVolumeClaim
    name: pvc-obs-static
    namespace: default
    resourceVersion: "819913"
    uid: 42fa0773-273f-4221-86d7-3198603c5998
  csi:
    driver: obs.csi.huaweicloud.com
    volumeHandle: custom-bucket
  persistentVolumeReclaimPolicy: Delete
  volumeMode: Filesystem
status:
  phase: Bound
```