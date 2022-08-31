# Using Block Volume

## Prerequisites

- kubernetes, EVS CSI Plugin

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/block/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/block/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/block/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                   READY   STATUS    RESTARTS   AGE
test-evs-block-nginx   1/1     Running   0          38s

```

```
# kubectl get pvc
NAME            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
evs-block-pvc   Bound    pvc-ffe70549-c77a-46be-9d11-df45063cffd2   10Gi       RWX            evs-sc         67s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                   STORAGECLASS   REASON   AGE
pvc-ffe70549-c77a-46be-9d11-df45063cffd2   10Gi       RWX            Delete           Bound    default/evs-block-pvc   evs-sc                  79s
```
