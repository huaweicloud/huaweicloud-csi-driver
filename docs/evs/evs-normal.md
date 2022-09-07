# Dynamic Provisioning

## Prerequisites

- kubernetes, EVS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/normal/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/normal/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/normal/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                    READY   STATUS    RESTARTS   AGE
test-evs-normal-nginx   1/1     Running   0          68s
```

```
# kubectl get pvc
NAME             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
evs-normal-pvc   Bound    pvc-b8c99f2b-8028-438b-8a0c-d6c493301d98   10Gi       RWX            evs-sc         26s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
pvc-b8c99f2b-8028-438b-8a0c-d6c493301d98   10Gi       RWX            Delete           Bound    default/evs-normal-pvc   evs-sc                  28s
```
