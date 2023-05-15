# Encrypted EVS (KMS)

## Prerequisites

- kubernetes, EVS CSI Driver

## How to use

### Step 1: Create SC with 

```
cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: evs-encryption-sc
provisioner: evs.csi.huaweicloud.com
allowVolumeExpansion: true
parameters:
  type: SSD
  kmsId: <the kms ID for disk encryption>
reclaimPolicy: Delete
EOF
```

### Step 2: Create PVC

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: evs-encryption-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: evs-encryption-sc
EOF
```

### Step 3: Create POD

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-evs-encryption-nginx
spec:
  containers:
    - image: nginx
      imagePullPolicy: IfNotPresent
      name: nginx
      ports:
        - containerPort: 80
          protocol: TCP
      volumeMounts:
        - mountPath: /var/lib/www/html
          name: encrypted-evs
  volumes:
    - name: encrypted-evs
      persistentVolumeClaim:
        claimName: evs-encryption-pvc
        readOnly: false
EOF
```

### Step 4: Check status of POD/PVC/PV

```
# kubectl get pod/test-evs-encryption-nginx
NAME                        READY   STATUS    RESTARTS   AGE
test-evs-encryption-nginx   1/1     Running   0          23s
```

```
# kubectl get pvc/evs-encryption-pvc
NAME                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS        AGE
evs-encryption-pvc   Bound    pvc-9dd98036-05e7-4be0-a43e-e119a2a2e91c   10Gi       RWX            evs-encryption-sc   68s
```

```
# kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS        REASON   AGE
pvc-9dd98036-05e7-4be0-a43e-e119a2a2e91c   10Gi       RWX            Delete           Bound    default/evs-encryption-pvc   evs-encryption-sc            76s
```

### Step 5: Clean up

```
kubectl delete pod/test-evs-encryption-nginx
kubectl delete pvc/evs-encryption-pvc
kubectl delete sc/evs-encryption-sc
```
