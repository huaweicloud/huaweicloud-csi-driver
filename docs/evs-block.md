## This document describes how to create a block store

## Step 1: create sc
```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/block/sc.yaml
```

## Step 2: create pvc
```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/block/pvc.yaml
```

## Step 3: create pod
```
kubectl create -f  https://raw.githubusercontent.com/huaweicloud/huaweicloud-csi-driver/master/examples/evs-csi-plugin/kubernetes/block/pod.yaml
```

## Step 4: check status
```
kubectl get all pod
```
The following output should be displayed:

```
NAME         READY   STATUS    RESTARTS   AGE
test-block   1/1     Running   0          96s
```

## Step 5: check resources in huawei cloud platform
Log in to Huawei cloud platform and check whether EVs resources are successfully created

## Step 6: delete pvc
```
kubectl delete pod test-block
```

## Step 7: check resources in huawei cloud platform
Log in to Huawei cloud platform and check whether EVs resources are successfully deleted
