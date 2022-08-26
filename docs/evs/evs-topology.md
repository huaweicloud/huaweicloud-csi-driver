# Topology

## Prerequisites

- kubernetes, EVS CSI Plugin

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/topology/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/topology/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/topology/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                      READY   STATUS    RESTARTS   AGE
test-evs-topology-nginx   1/1     Running   0          57s
```

```
# kubectl get pvc
NAME               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
evs-topology-pvc   Bound    pvc-3c32e26c-b570-4224-81c3-ba3e0a9dd99a   12Gi       RWO            topology-evs-sc   14s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                      STORAGECLASS      REASON   AGE
pvc-3c32e26c-b570-4224-81c3-ba3e0a9dd99a   12Gi       RWO            Delete           Bound    default/evs-topology-pvc   topology-evs-sc            5s
```
