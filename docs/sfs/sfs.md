# Huawei Cloud SFS CSI Plugin

SFS CSI Plugin provides the ability to interface with Huawei Cloud SFS storage resources.

Such as: create share volume, add access rules.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here
-https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| CSI version   | SFS CSI Plugin Version | Kubernetes Version Tested | Features                |
|---------------|------------------------|---------------------------|-------------------------|
| v1.5.0        | v0.1.0                 | v1.20 v1.21 v1.22 v1.23   | shareVolume accessRules |

## Deploy

### Prerequisites

- docker, kubeadm, kubelet and kubectl has been installed.

### Steps

- Create the config file

The driver initialization depends on a [cloud config file](../../deploy/cloud-config).
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

- Install Huawei Cloud SFS CSI Plugin

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-driver.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-node.yaml
```

- Waiting for all the pods in running

```
# kubectl get all -A
NAME                                   READY   STATUS    RESTARTS       AGE
...

```

## How to use

**Shares Volume:** [sample app](sfs-share.md)
