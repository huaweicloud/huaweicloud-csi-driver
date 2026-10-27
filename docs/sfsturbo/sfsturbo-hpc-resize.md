# HPC File System Expansion

## Overview

This example creates an SFS Turbo **HPC** file system (20MB/s/TiB) and expands it through the CSI driver. HPC capacity must follow the **1.2TiB-multiple** rule, so expansion also targets the next valid 1.2TiB step.

### How It Works

1. The user creates a StorageClass with `expandType=hpc`, `hpcBw` and `allowVolumeExpansion: true`.
2. The user creates a PVC requesting an initial HPC-valid capacity (e.g. 3686GiB = 3.6TiB).
3. To expand, the user updates the PVC requested storage to the next valid step (e.g. 4915GiB = 4.8TiB).
4. The CSI driver forwards the new capacity to the SFS Turbo API, which validates the 1.2TiB-multiple rule and performs the online expansion.

## Prerequisites

1. A Kubernetes cluster with the SFS Turbo CSI Driver deployed. See [SFS Turbo CSI Driver](sfsturbo.md) for deployment.
2. The StorageClass must enable expansion with `allowVolumeExpansion: true`.
3. The CSI `external-resizer` sidecar must be running.

## Usage

### 1. Create a StorageClass with expansion enabled

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: sfsturbo-hpc-sc
provisioner: sfsturbo.csi.huaweicloud.com
allowVolumeExpansion: true
reclaimPolicy: Delete
parameters:
  # shareType is required. For HPC file systems the SFS Turbo API does not
  # validate this field; STANDARD is fine.
  shareType: STANDARD
  # expandType=hpc creates an HPC file system (20MB/s/TiB ~ 1000MB/s/TiB).
  # When hpcBw is set, expandType can be omitted and is inferred as hpc.
  expandType: hpc
  # hpcBw is the HPC bandwidth spec. Valid values: 20M, 40M, 125M, 250M, 500M, 1000M.
  hpcBw: "20M"
  # autoCreateSecurityGroupRules is optional, should be 'true' or 'false'.
  # If not specified, the SFS Turbo API defaults to 'true'.
  autoCreateSecurityGroupRules: "false"
  # availability is optional; set it to override the AZ derived from topology.
  # availability: ap-southeast-1g
```

### 2. Create a PVC (3686Gi = 3.6TiB)

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-hpc-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      # HPC capacity must be a 1.2TiB multiple. 3686GiB = 3.6TiB, the minimum
      # valid size for 20MB/s/TiB.
      storage: 3686Gi
  storageClassName: sfsturbo-hpc-sc
```

### 3. Create a Pod that uses the PVC

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: sfsturbo-nginx-hpc-resize
spec:
  containers:
    - image: nginx
      name: sfsturbo-nginx-hpc-resize
      command: [ "/bin/sh" ]
      args: [ "-c", "while true; do echo $(date -u) >> /mnt/sfsturbo/outfile; sleep 5; done" ]
      volumeMounts:
        - mountPath: /mnt/sfsturbo
          name: sfsturbo-data
  volumes:
    - name: sfsturbo-data
      persistentVolumeClaim:
        claimName: sfsturbo-pvc-hpc-resize
```

### 4. Check the initial status

```shell
# kubectl get pod
NAME                        READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-hpc-resize   1/1     Running   0          5m1s

# kubectl get pvc
NAME                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
sfsturbo-pvc-hpc-resize     Bound    pvc-xxxx-xxxx-xxxx-xxxx                   3686Gi     RWX            sfsturbo-hpc-sc   5m26s

# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                              STORAGECLASS      AGE
pvc-xxxx-xxxx-xxxx-xxxx                    3686Gi     RWX            Delete           Bound    default/sfsturbo-pvc-hpc-resize    sfsturbo-hpc-sc   99s
```

### 5. Expand the volume to 4915Gi (4.8TiB)

Update the PVC requested storage to `4915Gi`:

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-hpc-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      # HPC capacity must be a 1.2TiB multiple. Expand to 4915GiB = 4.8TiB,
      # the next valid step after 3686GiB (3.6TiB).
      storage: 4915Gi
  storageClassName: sfsturbo-hpc-sc
```

```shell
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/hpc-resize/pvc2.yaml
```

### 6. Verify the new size on the running Pod

```shell
$ kubectl exec sfsturbo-nginx-hpc-resize -- df -h /mnt/sfsturbo
Filesystem       Size  Used Avail Use% Mounted on
192.168.0.175:/  4.8T     0  4.8T   0% /mnt/sfsturbo
```

## Notes

- HPC capacity **must** be a 1.2TiB multiple (e.g. 3686GiB = 3.6TiB, 4915GiB = 4.8TiB).
- The driver forwards the requested capacity to the SFS Turbo API without enforcing the legacy 500~32768 GiB range or the 100GiB expand step.
- The StorageClass must have `allowVolumeExpansion: true`; otherwise the resize is ignored.
- Expansion only **grows** the share; shrinking is not supported.
- See the [SFS Turbo API documentation](https://support.huaweicloud.com/intl/en-us/api-sfsturbo/CreateShare.html) for the per-spec size range and the 1.2TiB-multiple rule.
