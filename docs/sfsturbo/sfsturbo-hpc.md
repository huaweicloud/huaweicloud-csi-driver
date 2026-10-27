# Dynamic Provisioning (HPC file system)

This example creates an SFS Turbo **HPC** file system (20MB/s/TiB ~ 1000MB/s/TiB)
through the CSI driver using the `expandType=hpc` and `hpcBw` StorageClass parameters.

## Prerequisites

- kubernetes, SFS Turbo CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc/hpc-sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc/hpc-pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc/hpc-pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-hpc   1/1    Running   0          5m1s
```

```
# kubectl get pvc
NAME               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
sfsturbo-pvc-hpc   Bound    pvc-xxxx-xxxx-xxxx-xxxx                   3686Gi     RWX            sfsturbo-hpc-sc   5m26s
```

## Notes

- `hpcBw` selects the bandwidth spec: `20M`, `40M`, `125M`, `250M`, `500M`, `1000M`.
- The PVC capacity **must** be a 1.2TiB multiple (e.g. 3686GiB = 3.6TiB) and within the
  per-spec range. See the [SFS Turbo API documentation](https://support.huaweicloud.com/intl/en-us/api-sfsturbo/CreateShare.html).
- The driver forwards the requested capacity to the SFS Turbo API without enforcing the
  legacy 500~32768 GiB range or the 100GiB expand step.
