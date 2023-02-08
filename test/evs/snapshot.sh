#! /bin/sh
echo -e "====== Start Test EVS(snapshot) "

testRes="false"

cat << EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: evs-sc
provisioner: evs.csi.huaweicloud.com
allowVolumeExpansion: true
parameters:
  type: SSD
reclaimPolicy: Delete

---
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: evs-snapshot-class
driver: evs.csi.huaweicloud.com
deletionPolicy: Delete
parameters:
  force-create: "false"

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: evs-snapshot-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: evs-sc
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep evs-snapshot-pvc | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-snapshot-demo1
spec:
  containers:
    - image: nginx
      imagePullPolicy: IfNotPresent
      name: nginx
      command: ["/bin/sh"]
      args: ["-c", "while true; do echo $(date -u) >> /var/lib/www/html/outfile; sleep 1; done"]
      ports:
        - containerPort: 80
          protocol: TCP
      volumeMounts:
        - mountPath: /var/lib/www/html
          name: csi-data-evs
  volumes:
    - name: csi-data-evs
      persistentVolumeClaim:
        claimName: evs-snapshot-pvc
        readOnly: false
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-snapshot-demo1 | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl create -f -
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: new-snapshot-demo
spec:
  volumeSnapshotClassName: evs-snapshot-class
  source:
    persistentVolumeClaimName: evs-snapshot-pvc
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get VolumeSnapshot | grep new-snapshot-demo | grep true | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl create -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: snapshot-demo-restore
spec:
  storageClassName: evs-sc
  dataSource:
    name: new-snapshot-demo
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep snapshot-demo-restore | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-snapshot-demo2
spec:
  containers:
    - image: nginx
      imagePullPolicy: IfNotPresent
      name: nginx
      command: ["/bin/sh"]
      args: ["-c", "while true; do echo $(date -u) >> /var/lib/www/html/outfile; sleep 1; done"]
      ports:
        - containerPort: 80
          protocol: TCP
      volumeMounts:
        - mountPath: /var/lib/www/html
          name: csi-data-evs
  volumes:
    - name: csi-data-evs
      persistentVolumeClaim:
        claimName: snapshot-demo-restore
        readOnly: false
EOF


for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-snapshot-demo2 | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod test-snapshot-demo1
kubectl delete pod test-snapshot-demo2
kubectl delete pvc snapshot-demo-restore
kubectl delete pvc evs-snapshot-pvc
kubectl delete VolumeSnapshot new-snapshot-demo
kubectl delete VolumeSnapshotClass evs-snapshot-class
kubectl delete sc evs-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: EVS(snapshot) Test\n"
    exit 0
else
    echo -e "------ FAIL: EVS(snapshot) Test\n"
    exit 1
fi
