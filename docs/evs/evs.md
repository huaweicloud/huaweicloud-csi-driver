# Huawei Cloud EVS CSI Driver

The EVS CSI Driver is a CSI Specification compliant driver used by Container Orchestrators to manage
the lifecycle of Huawei Cloud EVS Volumes.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here 
-https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| EVS CSI Driver Version | CSI version | Kubernetes Version Tested | Features                |
|------------------------|-------------|---------------------------|-------------------------|
| v0.1.2                 | v1.5.0      | v1.20 v1.21 v1.22 v1.23   | volume resizer snapshot |

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

* `capacity` Optional. The EVS disk size. The value ranges from 10 GB to 32,768 GB. Defaults to 10 GB.
  It is located under `volumeAttributes`.

## Deploy

### Prerequisites

- docker, kubeadm, kubelet and kubectl has been installed.

### Steps

- Create the config file

Create the `cloud-config` file according to [cloud-config](../../deploy/evs-csi-plugin/cloud-config) in master node or control-plane,
see [Description of cloud config](../cloud-config.md) for configurations description.

See [IAM Policies for EVS CSI](../user-policy.md#iam-policies-for-evs-csi) for IAM policies.

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

**Dynamic Provisioning:** [evs normal](evs-normal.md)

**Volume Expansion:** [evs resize](evs-resize.md)

**Using Block Volume:** [evs block](evs-block.md)

**Volume Snapshots:** [evs snapshot](evs-snapshot.md)

**Ephemeral Volume:** [evs ephemeral](evs-ephemeral.md)

**Topology:** [evs topology](evs-topology.md)
