# Dynamic Provisioning

## Prerequisites

- kubernetes, SFS Turbo CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/dynamic/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/dynamic/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/dynamic/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                     READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-dynamic   1/1     Running   0          5m1s
```

```
# kubectl get pvc
NAME                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
sfsturbo-pvc-dynamic   Bound    pvc-a0aaab4f-e750-4821-8d18-30f27a9dcde3   500Gi      RWX            sfsturbo-sc    5m26s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   REASON   AGE
pvc-a0aaab4f-e750-4821-8d18-30f27a9dcde3   500Gi      RWX            Delete           Bound    default/sfsturbo-pvc-dynamic   sfsturbo-sc             99s
```
