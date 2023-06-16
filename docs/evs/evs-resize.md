# Volume Expansion

## Prerequisites

- kubernetes, EVS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/resize/sc.yaml
```

### Step 2: Create PVC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/resize/pvc.yaml
```

### Step 3: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/resize/pod.yaml
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                    READY   STATUS    RESTARTS   AGE
test-evs-resize-nginx   1/1     Running   0          63s

```

```
# kubectl get pvc
NAME                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
evs-normal-resize-pvc   Bound    pvc-54ed5cc9-928f-4dd2-98ab-d4da9c2ec5f6   10Gi       RWX            evs-sc         17s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   REASON   AGE
pvc-54ed5cc9-928f-4dd2-98ab-d4da9c2ec5f6   10Gi       RWX            Delete           Bound    default/evs-normal-resize-pvc   evs-sc                  16s
```

### Step 5: Resize volume storage

```
kubectl apply -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/resize/pvc2.yaml
```

### Step 6: Check filesystem size after resized on the running POD

```
$ kubectl exec test-evs-resize-nginx -- df -h /var/lib/www/html
Filesystem      Size  Used Avail Use% Mounted on
/dev/vdb         20G   44M   20G   1% /var/lib/www/html
```
