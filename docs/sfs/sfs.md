# Huawei Cloud SFS CSI Driver

The SFS CSI Driver is a CSI Specification compliant driver used by Container Orchestrators to manage
the lifecycle of Huawei Cloud SFS.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here
-https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| SFS CSI Driver Version | CSI version | Kubernetes Version Tested | Features |
|------------------------|-------------|---------------------------|----------|
| v0.1.1                 | v1.5.0      | v1.20 v1.21 v1.22 v1.23   | shares   |

## Deploy

### Prerequisites

- Kubernetes cluster

### Steps

- Create the config file

The driver initialization depends on a [cloud config file](../../deploy/sfs-csi-plugin/cloud-config).
Make sure it's in `/etc/sfs/cloud-config` on your master.

Click to view the detailed description of the file: [cloud-config](../cloud-config.md).

- Create secret resource

```
kubectl create secret -n kube-system generic cloud-config --from-file=/etc/sfs/cloud-config
```

This should create a secret name `cloud-config` in `kube-system` namespace.

Once the secret is created, Controller Plugin and Node Plugin can be deployed using respective manifests.

- Create RBAC resources

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/rbac-csi-sfs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/rbac-csi-sfs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/rbac-csi-sfs-secret.yaml
```

- Install Huawei Cloud SFS CSI driver

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-driver.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-node.yaml
```

- Waiting for all the PODs in running

```
# kubectl get all -A
NAMESPACE      NAME                                         READY   STATUS    RESTARTS   AGE
...
kube-system    pod/csi-sfs-controller-65f7488778-zmqpp      4/4     Running   0          50s
kube-system    pod/csi-sfs-node-9qckv                       3/3     Running   0          36s
kube-system    pod/csi-sfs-node-n5chb                       3/3     Running   0          36s
kube-system    pod/csi-sfs-node-wl4p4                       3/3     Running   0          36s
```

## Examples

**SFS Shares:** [share](sfs-share.md)
