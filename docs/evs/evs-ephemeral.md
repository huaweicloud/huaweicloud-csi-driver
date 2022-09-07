# Ephemeral Volume

## Prerequisites

- kubernetes, EVS CSI Driver

## How to use

### Step 1: Create SC

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/ephemeral/sc.yaml
```

### Step 2: Create POD

```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/ephemeral/pod.yaml
```

### Step 3: Check status of POD/PVC/PV

```
# kubectl get pod
NAME                      READY   STATUS    RESTARTS   AGE
ephemeral-example-nginx   1/1     Running   0          86s

```

```
# kubectl get pvc
NAME                                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS            AGE
ephemeral-example-nginx-scratch-volume   Bound    pvc-9bb98461-b58e-4ea2-86e4-e32c02124433   10Gi       RWO            scratch-storage-class   25s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                            STORAGECLASS            REASON   AGE
pvc-9bb98461-b58e-4ea2-86e4-e32c02124433   10Gi       RWO            Delete           Bound    default/ephemeral-example-nginx-scratch-volume   scratch-storage-class            32s
```
