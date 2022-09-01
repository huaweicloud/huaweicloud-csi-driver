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


```

```
# kubectl get pvc

```

```
# kubectl get pv

```
