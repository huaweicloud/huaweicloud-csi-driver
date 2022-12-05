#! /bin/sh
echo -e "====== Start Test EVS(resize) "

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
EOF

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: evs-normal-resize-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: evs-sc
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep evs-normal-resize-pvc | grep Bound | wc -l`
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
  name: test-evs-resize-nginx
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
          name: csi-data-evs
  volumes:
    - name: csi-data-evs
      persistentVolumeClaim:
        claimName: evs-normal-resize-pvc
        readOnly: false
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-evs-resize-nginx | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: evs-normal-resize-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 20Gi
  storageClassName: evs-sc
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep evs-normal-resize-pvc | grep Bound | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-evs-resize-nginx | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod test-evs-resize-nginx
kubectl delete pvc evs-normal-resize-pvc
kubectl delete sc evs-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: EVS(resize) Test\n"
    exit 0
else
    echo -e "------ FAIL: EVS(resize) Test\n"
    exit 1
fi
