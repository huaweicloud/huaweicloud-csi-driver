# SFS Turbo HPC File System (20MB/s/TiB ~ 1000MB/s/TiB)

## Overview

HPC file systems are high-performance SFS Turbo file systems with throughput specs from 20MB/s/TiB to 1000MB/s/TiB.
They are created through the CSI driver via the `expandType` and `hpcBw` StorageClass parameters, and are suitable for
high-throughput workloads such as AI training and big data analytics.

### How It Works

1. The user creates a StorageClass with `expandType=hpc` and `hpcBw=<spec>`.
2. The driver sets `metadata.expand_type=hpc` and `metadata.hpc_bw=<hpcBw>` on the create request.
3. The requested capacity is forwarded to the SFS Turbo API, which validates the per-spec size range and the 1.2TiB-multiple rule.
4. Unlike the traditional mode, the driver does **not** enforce the legacy 500~32768 GiB capacity range or the 100GiB expand step.

### Differences from the Traditional Mode

| Feature | Traditional Mode | HPC Mode                                                                       |
|---------|------------------|--------------------------------------------------------------------------------|
| File system type | Standard / Performance | 20MB/s/TiB, 40MB/s/TiB, 125MB/s/TiB, 250MB/s/TiB, 500MB/s/TiB and 1000MB/s/TiB |
| Enabled by | `shareType` | `expandType=hpc` + `hpcBw`                                                     |
| Capacity range | 500~32768 GiB, 100GiB step | Per-spec range, 1.2TiB multiple                                                |
| `shareType` validation | Validated by API | Not validated, can be `STANDARD`                                               |

## Prerequisites

1. A Kubernetes cluster with the SFS Turbo CSI Driver deployed. See [SFS Turbo CSI Driver](sfsturbo.md) for deployment.
2. The chosen `hpcBw` spec and requested capacity must satisfy the SFS Turbo HPC rules (see [Capacity Rules](#capacity-rules)).

## Usage

### 1. Create a StorageClass

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: sfsturbo-hpc-sc
provisioner: sfsturbo.csi.huaweicloud.com
reclaimPolicy: Delete
parameters:
  # shareType is required. For HPC file systems the SFS Turbo API does not
  # validate this field; STANDARD is fine.
  shareType: STANDARD
  # expandType=hpc creates an HPC file system.
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

### 2. Create a PVC

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-hpc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      # HPC capacity must be a 1.2TiB multiple (e.g. 3.6TiB -> 3686GiB) and
      # within the per-spec range. 20MB/s/TiB: 3686~1048576 GiB.
      storage: 3686Gi
  storageClassName: sfsturbo-hpc-sc
```

### 3. Create a Pod that uses the PVC

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: sfsturbo-nginx-hpc
spec:
  containers:
    - image: nginx
      name: sfsturbo-nginx-hpc
      command: [ "/bin/sh" ]
      args: [ "-c", "while true; do echo $(date -u) >> /mnt/sfsturbo/outfile; sleep 5; done" ]
      volumeMounts:
        - mountPath: /mnt/sfsturbo
          name: sfsturbo-data
  volumes:
    - name: sfsturbo-data
      persistentVolumeClaim:
        claimName: sfsturbo-pvc-hpc
```

### 4. Check status

```shell
# kubectl get pod
NAME                 READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-hpc   1/1     Running   0          5m1s

# kubectl get pvc
NAME               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
sfsturbo-pvc-hpc   Bound    pvc-xxxx-xxxx-xxxx-xxxx                   3686Gi     RWX            sfsturbo-hpc-sc   5m26s
```

## Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `shareType` | Yes | Should be `STANDARD` or `PERFORMANCE`. For HPC file systems this field is not validated by the SFS Turbo API and can be left as `STANDARD`. |
| `expandType` | No | Set to `hpc` to create an HPC file system. When `hpcBw` is set but `expandType` is omitted, the driver infers `expandType=hpc` automatically. |
| `hpcBw` | No | The HPC bandwidth spec. Valid values: `20M`, `40M`, `125M`, `250M`, `500M`, `1000M` (corresponding to 20MB/s/TiB ~ 1000MB/s/TiB). Required together with `expandType=hpc`. The value is not validated by the driver and is forwarded to the SFS Turbo API. |
| `autoCreateSecurityGroupRules` | No | Whether to automatically create security group rules for the share. `true` or `false`. If not specified, the SFS Turbo API defaults to `true`. |
| `availability` | No | Availability Zone (AZ) of the share. Overrides the AZ derived from topology. |

## Capacity Rules

HPC file system capacity **must** be a **1.2TiB multiple** (e.g. 3686GiB = 3.6TiB, 4915GiB = 4.8TiB) and within the per-spec size range.

Make sure the PVC requests a capacity that satisfies the HPC rules of the chosen `hpcBw` (see the [SFS Turbo API documentation](https://support.huaweicloud.com/intl/en-us/api-sfsturbo/CreateShare.html)).

| `hpcBw`         | Example size range |
|-----------------|---------------------|
| `20M`           | 3686 ~ 1048576 GiB |
| `40M` ~ `1000M` | See the SFS Turbo API documentation |

## Expansion

To expand an HPC file system, update the PVC requested storage. The new capacity must also be a 1.2TiB multiple and within the per-spec range. See [HPC File System Expansion](sfsturbo-hpc-resize.md) for a complete example.

## Limitations

1. The capacity must be a 1.2TiB multiple; otherwise the SFS Turbo API rejects the request.
2. The driver does not enforce the legacy 500~32768 GiB range or the 100GiB expand step — capacity rules are validated solely by the SFS Turbo API.
3. `hpcBw` is not validated by the driver; an invalid value is rejected by the SFS Turbo API at creation time.

