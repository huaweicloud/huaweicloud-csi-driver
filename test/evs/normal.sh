#! /bin/sh
echo -e "====== Start Test EVS(normal) "

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
  name: evs-normal-pvc
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
    lines=`kubectl get pvc | grep evs-normal-pvc | grep Bound | wc -l`
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
  name: test-evs-normal-nginx
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
        claimName: evs-normal-pvc
        readOnly: false
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-evs-normal-nginx | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod test-evs-normal-nginx
kubectl delete pvc evs-normal-pvc
kubectl delete sc evs-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: EVS(normal) Test\n"
    exit 0
else
    echo -e "------ FAIL: EVS(normal) Test\n"
    exit 1
fi
