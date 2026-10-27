# Dynamic Provisioning

## Overview

Dynamic provisioning creates a new SFS Turbo share on demand when a PVC is created.
The CSI driver reads the StorageClass parameters, provisions the share via the SFS Turbo API, and mounts it to the Pod automatically.
This is the default mode when no sub-directory or HPC annotations/parameters are provided.

### How It Works

1. The user creates a StorageClass with `shareType` set to `STANDARD` or `PERFORMANCE`.
2. The user creates a PVC referencing that StorageClass.
3. The CSI driver provisions a new SFS Turbo share via the SFS Turbo API.
4. On mount, the driver mounts the NFS export path into the Pod.

## Prerequisites

1. A Kubernetes cluster with the SFS Turbo CSI Driver deployed. See [SFS Turbo CSI Driver](sfsturbo.md) for deployment.
2. `reclaimPolicy: Delete` on the StorageClass (otherwise Kubernetes will not call `DeleteVolume`).

## Usage

### 1. Create a StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: sfsturbo-sc
provisioner: sfsturbo.csi.huaweicloud.com
reclaimPolicy: Delete
parameters:
  # shareType is required, should be 'STANDARD' or 'PERFORMANCE', defaults to 'STANDARD'
  shareType: STANDARD
  # autoCreateSecurityGroupRules is optional, should be 'true' or 'false'.
  # If not specified, the SFS Turbo API defaults to 'true'.
  # autoCreateSecurityGroupRules: "true"
```

### 2. Create a PVC

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-dynamic
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 500Gi
  storageClassName: sfsturbo-sc
```

### 3. Create a Pod that uses the PVC

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: sfsturbo-nginx-dynamic
spec:
  containers:
    - image: nginx
      name: sfsturbo-nginx-dynamic
      command: [ "/bin/sh" ]
      args: [ "-c", "while true; do echo $(date -u) >> /mnt/sfsturbo/outfile; sleep 5; done" ]
      volumeMounts:
        - mountPath: /mnt/sfsturbo
          name: sfsturbo-data
  volumes:
    - name: sfsturbo-data
      persistentVolumeClaim:
        claimName: sfsturbo-pvc-dynamic
```

### 4. Check status

```shell
# kubectl get pod
NAME                     READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-dynamic   1/1     Running   0          5m1s

# kubectl get pvc
NAME                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
sfsturbo-pvc-dynamic   Bound    pvc-a0aaab4f-e750-4821-8d18-30f27a9dcde3   500Gi      RWX            sfsturbo-sc    5m26s

# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-a0aaab4f-e750-4821-8d18-30f27a9dcde3   500Gi      RWX            Delete           Bound    default/sfsturbo-pvc-dynamic   sfsturbo-sc             99s
```

## Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `shareType` | Yes | `STANDARD` or `PERFORMANCE`. Selects the SFS Turbo file system type. |
| `autoCreateSecurityGroupRules` | No | Whether to automatically create security group rules for the share. `true` or `false`. If not specified, the SFS Turbo API defaults to `true`. |
| `availability` | No | Availability Zone (AZ) of the share. Overrides the AZ derived from topology. |

## Related Modes

- **HPC file system**: set `expandType=hpc` + `hpcBw` on the StorageClass. See [HPC file system](sfsturbo-hpc.md).
- **Sub-directory mode**: reuse an existing instance via PVC annotations. See [Sub-directory mode](sfsturbo-subdir-en.md).
- **Existing instance (static)**: bind a PV to an existing share by ID. See [Use an existing SFS Turbo](use-existing-sfsturbo.md).
