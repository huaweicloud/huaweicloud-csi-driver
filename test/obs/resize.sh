#! /bin/sh
echo -e "====== Start Test OBS(resize) "

testRes="false"

cat << EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: obs-sc
provisioner: obs.csi.huaweicloud.com
allowVolumeExpansion: true
reclaimPolicy: Delete
parameters:
  acl: public-read-write
EOF

cat << EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pvc-obs-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 50Gi
  storageClassName: obs-sc
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep pvc-obs-resize | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: nginx-obs-resize
spec:
  containers:
    - image: nginx
      name: nginx-obs-resize
      command: ["/bin/sh"]
      args: ["-c", "while true; do echo $(date -u) >> /mnt/obs/outfile; sleep 5; done"]
      volumeMounts:
        - mountPath: /mnt/obs
          name: obs-data
  volumes:
    - name: obs-data
      persistentVolumeClaim:
        claimName: pvc-obs-resize
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep nginx-obs-resize | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pvc-obs-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 60Gi
  storageClassName: obs-sc
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep pvc-obs-resize | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 5
done

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep nginx-obs-resize | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod nginx-obs-resize
kubectl delete pvc pvc-obs-resize
kubectl delete sc obs-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: OBS(resize) Test\n"
    exit 0
else
    echo -e "------ FAIL: OBS(resize) Test\n"
    exit 1
fi
