# SHare Expansion

## Prerequisites

- kubernetes, SFS Turbo CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/resize/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/resize/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/resize/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                    READY   STATUS    RESTARTS   AGE
sfsturbo-nginx-resize   1/1     Running   0          5m2s

```

```
# kubectl get pvc
NAME                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
sfsturbo-pvc-resize   Bound    pvc-bbbcc6f6-1380-4ac7-aa14-da59b4977ed2   500Gi      RWX            sfsturbo-sc    5m31s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                         STORAGECLASS   REASON   AGE
pvc-bbbcc6f6-1380-4ac7-aa14-da59b4977ed2   500Gi      RWX            Delete           Bound    default/sfsturbo-pvc-resize   sfsturbo-sc             108s
```

### Step 5: Resize volume storage

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfsturbo-csi-plugin/kubernetes/resize/pvc2.yaml
```

### Step 6: Check filesystem size after resized on the running POD

```
$ kubectl exec sfsturbo-nginx-resize -- df -h /mnt/sfsturbo
Filesystem       Size  Used Avail Use% Mounted on
192.168.0.175:/  600G     0  600G   0% /mnt/sfsturbo
```
