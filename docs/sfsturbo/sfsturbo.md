# Huawei Cloud SFS Turbo CSI Driver

The SFS Turbo CSI Driver is a CSI Specification compliant driver used by Container Orchestrators to manage
the lifecycle of Huawei Cloud SFS Turbo Shares.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here
-https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| CSI version | SFS Turbo CSI Driver Version | Kubernetes Version Tested | Features |
|-------------|------------------------------|---------------------------|----------|
| v1.5.0      | v0.1.0                       | v1.20 v1.21 v1.22 v1.23   | shares   |

## Supported Parameters

* `shareType` Required. Should be `STANDARD` or `PERFORMANCE` in SFS Turbo. It is located under `parameters`.

* `availability` Optional. Availability Zone(AZ) of the share. It is located under `parameters`.

## Deploy

### Prerequisites

- docker, kubeadm, kubelet and kubectl has been installed.

### Steps

- Create the config file

Create the `cloud-config` file according to [cloud-config](../../deploy/sfsturbo-csi-plugin/cloud-config) in master node or control-plane,
see [Description of cloud config](../cloud-config.md) for configurations description.

See [IAM Policies for SFS Turbo CSI](../iam-policies.md#iam-policies-for-sfs-turbo-csi) for IAM policies.

Use the following command create `cloud-config` secret:

```shell
kubectl create secret -n kube-system generic cloud-config --from-file=/etc/sfsturbo/cloud-config
```

- Create RBAC resources

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfsturbo-csi-plugin/kubernetes/rbac-csi-sfsturbo-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfsturbo-csi-plugin/kubernetes/rbac-csi-sfsturbo-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfsturbo-csi-plugin/kubernetes/rbac-csi-sfsturbo-secret.yaml
```

- Install HuaweiCloud EVS CSI Driver

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfsturbo-csi-plugin/kubernetes/csi-sfsturbo-driver.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfsturbo-csi-plugin/kubernetes/csi-sfsturbo-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfsturbo-csi-plugin/kubernetes/csi-sfsturbo-node.yaml
```

- Waiting for all the pods in running

```
# kubectl get all -A
NAMESPACE      NAME                                           READY   STATUS    RESTARTS       AGE
...
kube-system    pod/csi-sfsturbo-controller-56fcfbf7dc-r55n5   5/5     Running   0              5h33m
kube-system    pod/csi-sfsturbo-node-mwg6j                    3/3     Running   0              5h33m
```

## Examples

- [Automatically create SFS Turbo resources based on PVC](sfsturbo-dynamic.md)
- [Extending SFS Turbo resources bound to a PVC](sfsturbo-resize.md)
- [Use an existing SFS Turbo resource](use-existing-sfsturbo.md)
