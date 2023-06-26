# Huawei Cloud EVS CSI Driver

The EVS CSI Driver is a CSI Specification compliant driver used by Container Orchestrators to manage
the lifecycle of Huawei Cloud EVS Volumes.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here
-https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| EVS CSI Driver Version | CSI version | Kubernetes Version Tested | Features                |
|------------------------|-------------|---------------------------|-------------------------|
| v0.1.4                 | v1.5.0      | v1.20 v1.21 v1.22 v1.23   | volume resizer snapshot |
| v0.1.7                 | v1.5.0      | v1.20 ~ 1.25              | encryption              |
| v0.1.8                 | v1.5.0      | v1.20 ~ 1.25              |               |

> After `v0.1.8`, the API for querying ECS details has been upgraded,
> please use the latest IAM policies to modify your policy/role.
> 
> See [IAM Policies for EVS CSI](../iam-policies.md#iam-policies-for-evs-csi) for IAM policies.

## Supported Parameters

* `type` Required. Volume type, corresponding volume type should exist in EVS. It is located under `parameters`.

* `availability` Optional. Availability Zone(AZ) of the volume. It is located under `parameters`.

* `dssId` Optional. ID of the dedicated distributed storage used when creating a dedicated file system.
  It is located under `parameters`.

* `scsi` Optional. The device type of the EVS disks to be created. Defaults to `"false"`.
  It is located under `parameters`.
  - `"true"`:  the disk device type will be SCSI, which allows ECS OSs to directly access underlying storage media.
    SCSI reservation commands are supported.
  - `"false"`: the disk device type will be VBD, which supports only simple SCSI read/write commands.

* `kmsId` Optional. The KMS ID for disk encryption. If this parameter is specified, the disk will be encrypted.
  It is located under `parameters`.

> When a project first uses disk encryption, you need to create an agency that grants KMS access to EVS for every project in the region.

* `storage` Optional. The EVS disk size. The value ranges from 10 GB to 32,768 GB. Defaults to 10 GB.
  It is located under `volumeAttributes`.

### Example

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: evs-encryption
provisioner: evs.csi.huaweicloud.com
allowVolumeExpansion: true
parameters:
  type: SSD
  availability: ap-southeast-1a
  dssId: 591779b9-1863-48dc-b258-3a18b07212e5
  kmsId: 8f4e245e-1a46-4b14-a188-d8fe88211856
reclaimPolicy: Delete
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: evs-encryption-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: evs-encryption
```

## Deploy

### Prerequisites

- Kubernetes cluster
- [CSI Snapshotter](https://github.com/kubernetes-csi/external-snapshotter), if you don't use the volume snapshot
  feature,
  just ignore it.

### Steps

- Create the config file

Create the `cloud-config` file according to [cloud-config](../../deploy/evs-csi-plugin/cloud-config) in master node or control-plane,
see [Description of cloud config](../cloud-config.md) for configurations description.

See [IAM Policies for EVS CSI](../iam-policies.md#iam-policies-for-evs-csi) for IAM policies.

Use the following command create `cloud-config` secret:

```shell
kubectl create secret -n kube-system generic cloud-config --from-file=/etc/evs/cloud-config
```

- Create RBAC resources

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-secret.yaml
```

- Install HuaweiCloud EVS CSI Driver

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/csi-evs-driver.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/csi-evs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/csi-evs-node.yaml
```

- Waiting for all the pods in running

```
# kubectl get all -A
NAME                                   READY   STATUS    RESTARTS       AGE
...
csi-evs-plugin-bkkpb                   3/3     Running   0              3m22s
csi-evs-provisioner-54c44b746f-22p46   6/6     Running   0              88s
```

## Examples

The following are examples of specific functions:

**Dynamic Provisioning:** [dynamic-provisioning](dynamic-provisioning.md)

**Volume Expansion:** [evs resize](evs-resize.md)

**Using Block Volume:** [evs block](evs-block.md)

**Volume Snapshots:** [evs snapshot](evs-snapshot.md)

**Ephemeral Volume:** [evs ephemeral](evs-ephemeral.md)

**Topology:** [evs topology](evs-topology.md)

**Encryption:** [Encrypted EVS](evs-encrypted.md)
