# Huawei Cloud OBS CSI Driver

The OBS CSI Driver is a CSI Specification compliant driver used by Container Orchestrators to manage
the lifecycle of Huawei Cloud OBS Buckets.

## Compatibility

For sidecar version compatibility, please refer compatibility matrix for each sidecar here
https://kubernetes-csi.github.io/docs/sidecar-containers.html.

## Support version

| OBS CSI Driver Version | CSI version | Kubernetes Version Tested | Features        |
|------------------------|-------------|---------------------------|-----------------|
| v0.1.3                 | v1.5.0      | v1.20 v1.21 v1.22 v1.23   | buckets resizer |

> NOTE:
>
> OBS CSI will automatically install the obsfs tool on each node,
> and the following OS have been verified: CentOS 7, Ubuntu 16, HUAWEI CLOUD EulerOS 2 and EulerOS 2.
>
> If your Linux distribution is not Ubuntu 16 or CentOS 7, or a similar version,
> you need to configure the environment and execute the script,
> refer to [Generating obsfs by Compilation](https://support.huaweicloud.com/intl/en-us/fstg-obs/obs_12_0005.html).

## Supported Parameters

* `acl` Optional. Specifies the ACL policy for a bucket. The predefined common policies are as follows:
`private`, `public-read`, `public-read-write`, `public-read-delivered`, `public-read-write-delivered` and
`bucket-owner-full-control`. Defaults to `private`. It is located under `parameters`.

* `sseAlgorithm` Optional. Specifies the default encryption configuration for the bucket requires the use of a
  server-side encryption algorithm. Valid values are as follows: 
  - `kms`. Indicates that the server encryption is **SSE-KMS** mode.
  - `obs`. Indicates that the server encryption is **SSE-OBS** mode.

  default does not use encryption.

* `kmsKeyId` Optional. Specifies the KMS key id used in **SSE-KMS** encryption mode, if is not provided, the default key
  of KMS will be used.

* `projectId` Optional. Specifies the project ID to which the KMS key belongs under **SSE-KMS** encryption mode.
  - If **projectId** is provided, the **kmsKeyId** must belong to the **projectId**.
  - If **kmsKeyId** is not provided, the **projectId** cannot be provided.

## Deploy

### Prerequisites

- Kubernetes cluster

### Steps

- Create the config file

Create the `cloud-config` file according to [cloud-config](../../deploy/obs-csi-plugin/cloud-config) in master node or control-plane,
see [Description of cloud config](../cloud-config.md) for configurations description.

See [IAM Policies for OBS CSI](../iam-policies.md#iam-policies-for-obs-csi) for IAM policies.

Use the following command create `cloud-config` secret:

```shell
kubectl create secret -n kube-system generic cloud-config --from-file=/etc/obs/cloud-config
```

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

**Dynamic Provisioning With Encryption:** [obs encryption](obs-encryption.md)

**Use Existing Bucket:** [obs existing](obs-existing.md)

