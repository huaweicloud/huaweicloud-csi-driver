# Use existing SFS Turbo

This example describes how to provide an existing SFS Turbo resource for use by workloads.

## Prerequisites

Make sure that the SFS Turbo resource has been created under the region and project used by the Kubernetes node.

## Step1: Create a Persistent Volume (PV) with the existing SFS Turbo ID

After querying the SFS Turbo ID in the Huawei Cloud console, fill into the `volumeHandle` field below.

Use the following yaml to create a PV.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  namespace: default
  name: sfs-turbo-adf2
spec:
  capacity:
    storage: 500Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Delete
  csi:
    driver: sfsturbo.csi.huaweicloud.com
    volumeHandle: 66ced4ab-04d8-4dcc-97ea-684a48681a2b # The SFS Turbo ID wish to use.
```

Check the PV status.

```shell
$ kubectl get pv -n default | grep sfs-turbo-adf2
sfs-turbo-adf2      500Gi       RWX     Delete     Bound     default/pvc-sfs-turbo-adf2        15m
```

## Step2: Create a PVC and bind the PV created in the Step1

Use the following yaml to create a PVC.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  namespace: default
  name: pvc-sfs-turbo-adf2
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 500Gi
  volumeName: sfs-turbo-adf2 # This value is created in Step1.
```

Check that the status of the PVC is bound.

```shell
$ kubectl get pvc | grep pvc-sfs-turbo-adf2
pvc-sfs-turbo-adf2   Terminating   pvc-8059d095-2135-44cd-90d7-917783dd1a78   5Ti    RWX   obs-sc    1m
```

## Step3: Use PVC in the deployment

Use the following yaml to create a Pod with the PVC.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: default
  name: sfs-turbo-demo
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 1
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:1.7.9
          ports:
            - containerPort: 80
          command: ["/bin/sh"]
          args: ["-c", "while true; do echo $(date -u) >> /mnt/sfsturbo/outfile; sleep 5; done"]
          volumeMounts:
            - mountPath: /mnt/sfsturbo
              name: sfsturbo-data
      volumes:
        - name: sfsturbo-data
          persistentVolumeClaim:
            claimName: pvc-sfs-turbo-adf2 # This value is created in Step2.
```

Use the following command to check the status of the Pod.
When the status of the Pod is `Running`, it means that the SFS Turbo resource has been mounted and is working.

```shell
$ kubectl get pod sfs-turbo-demo
NAME                              READY   STATUS    RESTARTS   AGE
sfs-turbo-demo-689665698c-vm75n   1/1     Running   1          1m
```
