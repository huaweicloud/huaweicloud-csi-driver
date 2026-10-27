# Share Expansion

## Overview

Share expansion resizes an SFS Turbo share in place when the PVC requested storage increases. The CSI driver forwards the new capacity to the SFS Turbo API, which performs the online expansion. Expansion only grows the share; shrinking is not supported.

> For **HPC file systems**, capacity must follow the 1.2TiB-multiple rule. See [HPC File System Expansion](sfsturbo-hpc-resize.md) instead.

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
  name: sfsturbo-sc
provisioner: sfsturbo.csi.huaweicloud.com
allowVolumeExpansion: true
reclaimPolicy: Delete
parameters:
  # shareType is required, should be 'STANDARD' or 'PERFORMANCE', defaults to 'STANDARD'
  shareType: STANDARD
```

### 2. Create a PVC (500Gi)

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-resize
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
  name: sfsturbo-nginx-resize
spec:
  containers:
    - image: nginx
      name: sfsturbo-nginx-resize
      command: [ "/bin/sh" ]
      args: [ "-c", "while true; do echo $(date -u) >> /mnt/sfsturbo/outfile; sleep 5; done" ]
      volumeMounts:
        - mountPath: /mnt/sfsturbo
          name: sfsturbo-data
  volumes:
    - name: sfsturbo-data
      persistentVolumeClaim:
        claimName: sfsturbo-pvc-resize
```

### 4. Check the initial status

```shell
# kubectl get pod
NAME                    READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-resize   1/1     Running   0          5m2s

# kubectl get pvc
NAME                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
sfsturbo-pvc-resize   Bound    pvc-bbbcc6f6-1380-4ac7-aa14-da59b4977ed2   500Gi      RWX            sfsturbo-sc    5m31s

# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                         STORAGECLASS   REASON   AGE
pvc-bbbcc6f6-1380-4ac7-aa14-da59b4977ed2   500Gi      RWX            Delete           Bound    default/sfsturbo-pvc-resize   sfsturbo-sc             108s
```

### 5. Expand the volume to 600Gi

Update the PVC requested storage to `600Gi`:

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 600Gi
  storageClassName: sfsturbo-sc
```

```shell
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/resize/pvc2.yaml
```

### 6. Verify the new size on the running Pod

```shell
$ kubectl exec sfsturbo-nginx-resize -- df -h /mnt/sfsturbo
Filesystem       Size  Used Avail Use% Mounted on
192.168.0.175:/  600G     0  600G   0% /mnt/sfsturbo
```

## Notes

- Expansion only **grows** the share; shrinking is not supported.
- The StorageClass must have `allowVolumeExpansion: true`; otherwise the resize is ignored.
- The new capacity must satisfy the SFS Turbo per-spec size range and the expand step (100Gi for Standard/Performance).
- For HPC file systems, capacity must be a 1.2TiB multiple. See [HPC File System Expansion](sfsturbo-hpc-resize.md).
