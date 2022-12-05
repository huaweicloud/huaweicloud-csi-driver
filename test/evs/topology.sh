#! /bin/sh
echo -e "====== Start Test EVS(topology) "

testRes="false"

cat << EOF | kubectl create -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: topology-evs-sc
provisioner: evs.csi.huaweicloud.com
volumeBindingMode: WaitForFirstConsumer
allowedTopologies:
  - matchLabelExpressions:
      - key: topology.evs.csi.huaweicloud.com/zone
        values:
          - ap-southeast-1b
parameters:
  type: SSD
EOF

cat << EOF | kubectl create -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: evs-topology-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 12Gi
  storageClassName: topology-evs-sc
EOF

for (( i=0; i<4; i++));
do
    lines=`kubectl get pvc | grep evs-topology-pvc | grep Bound | wc -l`
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
  name: test-evs-topology-nginx
spec:
  containers:
    - image: nginx
      imagePullPolicy: IfNotPresent
      name: nginx
      ports:
        - containerPort: 80
          protocol: TCP
      volumeMounts:
        - mountPath: /var/lib/www/data
          name: data
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: evs-topology-pvc
        readOnly: false
EOF

for (( i=0; i<10; i++));
do
    lines=`kubectl get pod | grep test-evs-topology-nginx | grep Running | wc -l`
    if [ "$lines" = "1" ]; then
        testRes="true"
    else
        testRes="false"
    fi
    sleep 5
done

kubectl delete pod test-evs-topology-nginx
kubectl delete pvc evs-topology-pvc
kubectl delete sc topology-evs-sc

if [ "$testRes" = "true" ]; then
    echo -e "------ PASS: EVS(topology) Test\n"
    exit 0
else
    echo -e "------ FAIL: EVS(topology) Test\n"
    exit 1
fi
