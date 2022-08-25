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

## Step 4: check status of pod/pvc/pv
```
# kubectl describe pod/test-block
Events:
  Type     Reason                  Age                    From                     Message
  ----     ------                  ----                   ----                     -------
  Warning  FailedScheduling        6m50s (x2 over 6m53s)  default-scheduler        0/1 nodes are available: 1 pod has unbound immediate PersistentVolumeClaims.
  Normal   Scheduled               6m47s                  default-scheduler        Successfully assigned default/test-block to k8s-lu
  Normal   SuccessfulAttachVolume  6m28s                  attachdetach-controller  AttachVolume.Attach succeeded for volume "pvc-00e1dd22-152e-4d77-89dd-8014661023f3"
  Normal   SuccessfulMountVolume   6m13s                  kubelet                  MapVolume.MapPodDevice succeeded for volume "pvc-00e1dd22-152e-4d77-89dd-8014661023f3" globalMapPath "/var/lib/kubelet/plugins/kubernetes.io/csi/volumeDevices/pvc-00e1dd22-152e-4d77-89dd-8014661023f3/dev"
  Normal   SuccessfulMountVolume   6m13s                  kubelet                  MapVolume.MapPodDevice succeeded for volume "pvc-00e1dd22-152e-4d77-89dd-8014661023f3" volumeMapPath "/var/lib/kubelet/pods/0fd34ed2-ad1f-4edb-9b5b-ae87ad0988ed/volumeDevices/kubernetes.io~csi"
  Normal   Pulled                  6m12s                  kubelet                  Container image "nginx" already present on machine
  Normal   Created                 6m12s                  kubelet                  Created container nginx
  Normal   Started                 6m12s                  kubelet                  Started container nginx
```

```
# kubectl get pvc
NAME                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
csi-pvc-evs-block   Bound    pvc-00e1dd22-152e-4d77-89dd-8014661023f3   10Gi       RWX            evs-sc         8m29s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                       STORAGECLASS   REASON   AGE
pvc-00e1dd22-152e-4d77-89dd-8014661023f3   10Gi       RWX            Delete           Bound    default/csi-pvc-evs-block   evs-sc                  8m53s
```

## Step 5: check resources in huawei cloud platform
Log in to Huawei cloud platform and check whether EVs resources are successfully created

## Step 6: delete pod
```
kubectl delete pod test-block
```

## Step 7: delete pvc
```
kubectl delete pvc csi-pvc-evs-block
```

## Step 8: check resources in huawei cloud platform
Log in to Huawei cloud platform and check whether EVs resources are successfully deleted
