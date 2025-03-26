# Dynamic Provisioning

Dynamic Provisioning will create a new bucket and set bucket encryption. When removing SC or PVC, the bucket resources
will be cleared and deleted together.

## Prerequisites

- kubernetes, OBS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/encryption/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/encryption/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/encryption/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
obs-encryption-kms   1/1     Running   0          19s
```

```
# kubectl get pvc
NAME                 STATUS   VOLUME                                     CAPACITY     ACCESS MODES   STORAGECLASS               AGE
obs-encryption-kms   Bound    pvc-aa02b032-a2dd-4a07-bcf5-0515f79fc0d2   500Gi        RWX            obs-encryption-kms         52s
```

```
# kubectl get pv
NAME                                       CAPACITY     ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS               REASON   AGE
pvc-aa02b032-a2dd-4a07-bcf5-0515f79fc0d2   500Gi        RWX            Delete           Bound    default/obs-encryption-kms   obs-encryption-kms                  101s
```
