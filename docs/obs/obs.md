# Huawei Cloud OBS CSI Driver

The OBS CSI Driver is a CSI Specification compliant driver used by Container Orchestrators to manage
the lifecycle of Huawei Cloud OBS Buckets.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here
https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| OBS CSI Driver Version | CSI version | Kubernetes Version Tested | Features        |
|------------------------|-------------|---------------------------|-----------------|
| v0.1.0                 | v1.5.0      | v1.20 v1.21 v1.22 v1.23   | buckets resizer |

## Supported Parameters

* `acl` Optional. Specifies the ACL policy for a bucket. The predefined common policies are as follows:
`private`, `public-read`, `public-read-write`, `public-read-delivered`, `public-read-write-delivered` and
`bucket-owner-full-control`. Defaults to `private`. It is located under `parameters`.

## Deploy

### Prerequisites

- docker, kubeadm, kubelet and kubectl has been installed.

### Steps

- Create the config file

Create the `cloud-config` file according to [cloud config file](../../deploy/obs-csi-plugin/cloud-config) 
in the master node.

See [cloud-config](../cloud-config.md) for configurations description.

- Create secret resource

```
kubectl create secret -n kube-system generic cloud-config --from-file=/etc/obs/cloud-config
```

This will create `cloud-config` secret in `kube-system` namespace.

Once the secret is created, Controller Plugin and Node Plugin can be deployed using respective manifests.

- Create RBAC resources

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/obs-csi-plugin/kubernetes/rbac-csi-obs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/obs-csi-plugin/kubernetes/rbac-csi-obs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/obs-csi-plugin/kubernetes/rbac-csi-obs-secret.yaml
```

- Install HuaweiCloud OBS CSI Driver

```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/obs-csi-plugin/kubernetes/csi-obs-driver.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/obs-csi-plugin/kubernetes/csi-obs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/obs-csi-plugin/kubernetes/csi-obs-node.yaml
```

- Waiting for all the Pods in running

```
# kubectl get all -A
NAMESPACE      NAME                                                 READY   STATUS    RESTARTS       AGE
...
kube-system    pod/csi-obs-controller-687dc77b4d-lm54p              5/5     Running   5 (167m ago)   3h30m
kube-system    pod/csi-obs-plugin-fvswm                             3/3     Running   3 (167m ago)   3h30m
kube-system    pod/csi-obs-plugin-l6d5b                             3/3     Running   3 (167m ago)   3h30m
```

## Examples

The following are examples of specific functions:

**Dynamic Provisioning:** [obs dynamic](obs-dynamic.md)

**Use Existing Bucket:** [obs existing](obs-existing.md)
