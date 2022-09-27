#! /bin/sh
echo -e "====== Start Test SFS Turbo(resize) "

testRes="false"

cat << EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: sfsturbo-sc
provisioner: sfsturbo.csi.huaweicloud.com
allowVolumeExpansion: true
reclaimPolicy: Delete
parameters:
  # shareType is required, should be 'STANDARD' or 'PERFORMANCE', defaults to 'STANDARD'
  shareType: STANDARD
EOF

cat << EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 500Gi
  storageClassName: sfsturbo-sc
EOF

for (( i=0; i<20; i++));
do
    lines=`kubectl get pvc | grep sfsturbo-pvc-resize | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 20
done

cat << EOF | kubectl create -f -
apiVersion: v1
kind: Pod
metadata:
  name: sfsturbo-nginx-resize
spec:
  containers:
    - image: nginx
      name: sfsturbo-nginx-resize
      command: [ "/bin/sh" ]
      args: [ "-c", "while true; do echo $(date -u) >> /mnt/sfsturbo/outfile; sleep 5; done" ]
      volumeMounts:
        - mountPath: /mnt/sfsturbo
          name: sfsturbo-data
  volumes:
    - name: sfsturbo-data
      persistentVolumeClaim:
        claimName: sfsturbo-pvc-resize
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep sfsturbo-nginx-resize | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 10
done

cat << EOF | kubectl apply -f -
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: sfsturbo-pvc-resize
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 600Gi
  storageClassName: sfsturbo-sc
EOF

for (( i=0; i<20; i++));
do
    lines=`kubectl get pvc | grep sfsturbo-pvc-resize | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 20
done

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep sfsturbo-nginx-resize | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
        break
    else
        testRes="false"
    fi
    sleep 10
done

kubectl delete pod sfsturbo-nginx-resize
kubectl delete pvc sfsturbo-pvc-resize
kubectl delete sc sfsturbo-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: SFS Turbo(resize) Test\n"
    exit 0
else
    echo -e "------ FAIL: SFS Turbo(resize) Test\n"
    exit 1
fi
