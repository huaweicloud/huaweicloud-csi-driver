
## Huawei Cloud EVS-CSI Plugin Introduction

EVS disk types and features are provided to help you quickly select your desired disk type to match services. This facilitates rapid service migration to HUAWEI CLOUD. 

After purchasing a data disk, you must attach the disk to a server and initialize (format) it before using it. You can expand, detach, or delete your disk any time based on service requirements; easily manage your disks and snapshots; and view monitoring metrics in real time to learn about the health status of your disks.

Abundant EVS APIs and calling examples help you manage EVS resources, including disks, snapshots, tags, and transfers.

## EVS-CSI Features And Available Versions

| EVS Plugin                           | Tag      | Kubernetes Version   | Features                       |
|--------------------------------------|----------|----------------------|--------------------------------|
| docker.io/huaweicloud/evs-csi-plugin | v1.0.0   | 1.18.0               | volume attach resizer snapshot |
| docker.io/huaweicloud/evs-csi-plugin | xxx      | 1.17.0               | volume snapshot                |
| docker.io/huaweicloud/evs-csi-plugin | xxx      | 1.17.0               | volume snapshot                |

## Compiling and Package

You can directly use the plug-in image provided in the table. It is recommended to use the latest version of the plug-in.

Or use your own packaged image. for example,you can package the project into an image with docker and push it to dockerhub,
then replace "evs-csi-provisioner" with your image address in "csi-evs-controller.yaml"

## Preparation

Apply for a cloud server in Huawei cloud, and prepare kubernetes environment

The driver initialization depends on a [cloud config file](../deploy/evs-csi-plugin/cloud-config). Make sure it's in `/etc/evs/cloud-config` on your node.
The following is an explanation of the file fields
```
[Global]
cloud=          # myhuaweicloud.com
auth-url=       # Auth url 
region=         # Region where you control resourses
access-key=     # AK Infomation
secret-key=     # SK Infomation
project-id=     # ProjectID in your region
[Vpc]
id=             # VPC where your cluster resides
```

## Use steps

### Step 1: create rbac csi
```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/rbac-csi-evs-secret.yaml
```

### Step 2: create controller csi、driver csi、node csi
```
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/csi-evs-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/csi-evs-node.yaml
kubectl apply -f https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/deploy/evs-csi-plugin/kubernetes/csi-evs-secret.yaml
```
### Step 3: check csi plugin status
```
kubectl get all -A
```
The following output should be displayed:

```
NAME                                   READY   STATUS    RESTARTS       AGE
csi-evs-plugin-bkkpb                   3/3     Running   0              3m22s
csi-evs-provisioner-54c44b746f-22p46   6/6     Running   0              88s
```

### Step4: create resources

The following is the example of specific functions:

**EVS Block:** [evs-block](./evs-block.md)
**EVS EPHEMERAL:** [evs-ephemeral](./evs-ephemeral.md)
**EVS NORMAL:** [evs-normal](./evs-normal.md)
**EVS RESIZE:** [evs-resize](./evs-resize.md)
**EVS SNAPSHOT:** [evs-snapshot](./evs-snapshot.md)
**EVS TOPOLOGY:** [evs-topology](./evs-topology.md)

