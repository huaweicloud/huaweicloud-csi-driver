# Shared File System Volume

## Prerequisites

- kubernetes, SFS CSI Plugin

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfs-csi-plugin/kubernetes/share/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfs-csi-plugin/kubernetes/share/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/sfs-csi-plugin/kubernetes/share/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME        READY   STATUS    RESTARTS   AGE
nginx-sfs   1/1     Running   0          81s
```

```
# kubectl get pvc
NAME      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-sfs   Bound    pvc-9278db4d-cb8d-4ee8-a85b-04ca355987b8   10Gi       RWX            sfs-sc         45s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS   REASON   AGE
pvc-9278db4d-cb8d-4ee8-a85b-04ca355987b8   10Gi       RWX            Delete           Bound    default/pvc-sfs   sfs-sc                  40s
```
