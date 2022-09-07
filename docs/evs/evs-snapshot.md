# Snapshot Create and Restore

## Prerequisites

- kubernetes, EVS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/snapshot/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/snapshot/pvc.yaml
```

### Step 3: Check status of PVC/PV

```
# kubectl get pvc
NAME               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
evs-snapshot-pvc   Bound    pvc-e164374c-eb41-4eb6-951e-1194f141058f   10Gi       RWO            evs-sc         34s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                      STORAGECLASS   REASON   AGE
pvc-e164374c-eb41-4eb6-951e-1194f141058f   10Gi       RWO            Delete           Bound    default/evs-snapshot-pvc   evs-sc                  52s
```

### Step 4: Create VolumeSnapshotClass

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/snapshot/snapshot-class.yaml
```

### Step 5: Create snapshot

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/snapshot/snapshot-create.yaml
```

### Step 6: Restore PVC by snapshot

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/snapshot/snapshot-restore.yaml
```

### Step 7: Check restore PVC

```
# kubectl get pvc
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
evs-snapshot-pvc        Bound    pvc-e164374c-eb41-4eb6-951e-1194f141058f   10Gi       RWO            evs-sc         2m45s
snapshot-demo-restore   Bound    pvc-4816cf76-2e6a-4722-a33d-bfc14719d673   10Gi       RWO            evs-sc         24s
```
