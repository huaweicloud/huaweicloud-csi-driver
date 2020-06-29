# SFS CSI Driver for Kubernetes
Huawei Cloud SFS Container Storage Interface (CSI) Plugin `sfs.csi.huaweicloud.com`

### Prerequisite
 - The driver initialization depends on a [cloud config file](./deploy/cloud-config). Make sure it's in `/etc/sfs/cloud-config` on your node.

### Install SFS CSI driver

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/rbac-csi-sfs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/rbac-csi-sfs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/sfs-csi-plugin/kubernetes/csi-sfs-driver.yaml
```

### Examples

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfs-csi-plugin/kubernetes/sc.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfs-csi-plugin/kubernetes/pvc.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfs-csi-plugin/kubernetes/pod.yaml
```

### Links
 - [Kubernetes CSI Documentation](https://kubernetes-csi.github.io/docs/Home.html)
 - [CSI Drivers](https://github.com/kubernetes-csi/drivers)
 - [Container Storage Interface (CSI) Specification](https://github.com/container-storage-interface/spec)
