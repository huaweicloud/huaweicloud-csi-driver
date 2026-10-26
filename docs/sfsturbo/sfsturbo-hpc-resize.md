# HPC File System Expansion

This example creates an SFS Turbo **HPC** file system (20MB/s/TiB) and expands it
through the CSI driver.

## Prerequisites

- kubernetes, SFS Turbo CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc-resize/sc.yaml
```

### Step 2: Create PVC (3686Gi = 3.6TiB)

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc-resize/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc-resize/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                        READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-hpc-resize   1/1    Running   0          5m1s
```

```
# kubectl get pvc
NAME                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
sfsturbo-pvc-hpc-resize     Bound    pvc-xxxx-xxxx-xxxx-xxxx                   3686Gi     RWX            sfsturbo-hpc-sc   5m26s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                              STORAGECLASS      AGE
pvc-xxxx-xxxx-xxxx-xxxx                    3686Gi     RWX            Delete           Bound    default/sfsturbo-pvc-hpc-resize    sfsturbo-hpc-sc   99s
```

### Step 5: Resize volume storage to 4915Gi (4.8TiB)

```
kubectl apply -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc-resize/pvc2.yaml
```

### Step 6: Check filesystem size after resized on the running POD

```
$ kubectl exec sfsturbo-nginx-hpc-resize -- df -h /mnt/sfsturbo
Filesystem       Size  Used Avail Use% Mounted on
192.168.0.175:/  4.8T     0  4.8T   0% /mnt/sfsturbo
```

## Notes

- HPC capacity **must** be a 1.2TiB multiple (e.g. 3686GiB = 3.6TiB, 4915GiB = 4.8TiB).
- The driver forwards the requested capacity to the SFS Turbo API without enforcing the
  legacy 500~32768 GiB range or the 100GiB expand step.
- See the [SFS Turbo API documentation](https://support.huaweicloud.com/intl/en-us/api-sfsturbo/CreateShare.html)
  for the per-spec size range and the 1.2TiB-multiple rule.
