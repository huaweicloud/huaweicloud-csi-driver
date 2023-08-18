# GPSSD2 Volume

## Prerequisites

- kubernetes, EVS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/gpssd2-volume/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/gpssd2-volume/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/gpssd2-volume/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                   READY   STATUS    RESTARTS   AGE
test-evs-gpssd2-nginx  1/1     Running   0          41s

```

```
# kubectl get pvc
NAME             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
evs-gpssd2-pvc   Bound    pvc-9a2084fb-03fc-46f0-970d-fc4cd9c7b212   15Gi       RWX            evs-sc-gpssd2  91s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
pvc-9a2084fb-03fc-46f0-970d-fc4cd9c7b212   10Gi       RWX            Delete           Bound    default/evs-gpssd2-pvc   evs-sc-gpssd2           92s
```
