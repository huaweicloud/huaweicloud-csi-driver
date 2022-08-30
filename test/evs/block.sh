#! /bin/sh
echo -e "======== Start Test EVS(block)"

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
  name: evs-block-pvc
spec:
  accessModes:
    - ReadWriteMany
  volumeMode: Block
  resources:
    requests:
      storage: 10Gi
  storageClassName: evs-sc
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pvc | grep evs-block-pvc | grep Bound | wc -l`
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
  name: test-evs-block-nginx
spec:
  containers:
    - image: nginx
      imagePullPolicy: IfNotPresent
      name: nginx
      ports:
        - containerPort: 80
          protocol: TCP
      volumeDevices:
        - devicePath: /var/lib/www/html
          name: csi-data
  volumes:
    - name: csi-data
      persistentVolumeClaim:
        claimName: evs-block-pvc
        readOnly: false
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-evs-block-nginx | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod test-evs-block-nginx
kubectl delete pvc evs-block-pvc
kubectl delete sc evs-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: EVS Test(block)\n"
    exit 0
else
    echo -e "------ FAIL: EVS Test(block)\n"
    exit 1
fi
