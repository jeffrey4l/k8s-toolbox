ceph osd pool create kube 1024

ceph auth get-or-create client.kube mon 'allow r' osd 'allow class-read object_prefix rbd_children, allow rwx pool=kube'

cat <<EOF | oc -n kube-system apply -f -
---
apiVersion: v1
kind: Secret
metadata:
  name: ceph-secret
data:
  key: $(ceph auth get-key client.admin | base64 -w0)
type: kubernetes.io/rbd
EOF

cat <<EOF | oc apply -f -
---
apiVersion: v1
kind: Secret
metadata:
  name: ceph-user-secret
data:
  key: $(ceph auth get-key client.kube | base64 -w0)
type: kubernetes.io/rbd
EOF

cat <<EOF | oc apply -f -
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ceph-default
  annotations:
    storageclass.beta.kubernetes.io/is-default-class: "true"
provisioner: kubernetes.io/rbd
parameters:
  monitors: 10.3.0.21:6789,10.3.0.23:6789,10.3.0.24:6789
  adminId: admin
  adminSecretName: ceph-secret
  adminSecretNamespace: kube-system
  pool: kube
  userId: kube
  userSecretName: ceph-user-secret
EOF

## example

cat <<EOF | oc apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: centos7
spec:
  containers:
    - name: centos7
      image: centos:7
      command: ["sleep", "60000"]
      volumeMounts:
        - name: vol1
          mountPath: /data
          readOnly: false
  volumes:
    - name: vol1
      persistentVolumeClaim:
        claimName: test
EOF

cat <<EOF | oc apply -f -
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: centos7
spec:
  selector:
    matchLabels:
        app: centos7
  template:
    metadata:
      labels:
        app: centos7
    spec:
      containers:
        - name: centos7
          image: centos:7
          command: ["sleep", "60000"]
          volumeMounts:
            - name: vol1
              mountPath: /data
              readOnly: false
      volumes:
        - name: vol1
          persistentVolumeClaim:
            claimName: test
EOF
