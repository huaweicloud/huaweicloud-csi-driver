# Dynamic Provisioning

Dynamic Provisioning will create a new bucket. When removing SC or PVC, the bucket resources will be cleared and deleted together.

## Prerequisites

- kubernetes, OBS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/dynamic/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/dynamic/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/obs-csi-plugin/kubernetes/dynamic/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME        READY   STATUS    RESTARTS   AGE
nginx-obs   1/1     Running   0          13s
```

```
# kubectl get pvc
NAME      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-obs   Bound    pvc-aa02b032-a2dd-4a07-bcf5-0515f79fc0d2   5Gi        RWX            obs-sc         46s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS   REASON   AGE
pvc-aa02b032-a2dd-4a07-bcf5-0515f79fc0d2   5Gi        RWX            Delete           Bound    default/pvc-obs   obs-sc                  86s
```
