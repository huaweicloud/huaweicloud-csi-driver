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

The driver initialization depends on a [cloud config file](../../deploy/sfsturbo-csi-plugin/cloud-config).
Make sure it's in `/etc/sfsturbo/cloud-config` on your master.

Click to view the detailed description of the file: [cloud-config](../cloud-config.md).

- Create secret resource

```
kubectl create secret -n kube-system generic cloud-config --from-file=/etc/sfsturbo/cloud-config
```

This should create a secret name `cloud-config` in `kube-system` namespace.

Once the secret is created, Controller Plugin and Node Plugin can be deployed using respective manifests.

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
