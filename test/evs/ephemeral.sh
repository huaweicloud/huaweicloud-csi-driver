#! /bin/sh
echo -e "====== Start Test EVS(ephemeral) "

testRes="false"

cat << EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: scratch-storage-class
provisioner: evs.csi.huaweicloud.com
parameters:
  type: SSD
EOF

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ephemeral-example-nginx
spec:
  containers:
  - image: nginx
    imagePullPolicy: IfNotPresent
    name: nginx-inline
    volumeMounts:
    - name: scratch-volume
      mountPath: /var/lib/www/html
  volumes:
  - name: scratch-volume
    ephemeral:
      volumeClaimTemplate:
        metadata:
          labels:
            type: my-frontend-volume
        spec:
          accessModes: [ "ReadWriteOnce" ]
          storageClassName: scratch-storage-class
          resources:
            requests:
              storage: 10Gi
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep ephemeral-example-nginx | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod ephemeral-example-nginx
kubectl delete sc scratch-storage-class

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: EVS(ephemeral) Test\n"
    exit 0
else
    echo -e "------ FAIL: EVS(ephemeral) Test\n"
    exit 1
fi
