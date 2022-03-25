package constants

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

const (
	TunnelServerClusterRole = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  name: tunnel-server
rules:
- apiGroups:
    - certificates.k8s.io
  resources:
    - certificatesigningrequests
  verbs:
    - create
    - get
    - list
    - watch
- apiGroups:
    - certificates.k8s.io
  resources:
    - certificatesigningrequests/approval
  verbs:
    - update
- apiGroups:
    - certificates.k8s.io
  resourceNames:
    - kubernetes.io/legacy-unknown
  resources:
    - signers
  verbs:
    - approve
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - list
  - watch
  - get
  - create
  - update
- apiGroups:
  - "coordination.k8s.io"
  resources:
  - leases
  verbs:
  - create
  - get
  - update
`
	TunnelServerServiceAccount = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tunnel-server
  namespace: kube-system
`
	TunnelServerClusterRolebinding = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tunnel-server
subjects:
  - kind: ServiceAccount
    name: tunnel-server
    namespace: kube-system
roleRef:
  kind: ClusterRole
  name: tunnel-server
  apiGroup: rbac.authorization.k8s.io
`
	TunnelServerService = `
apiVersion: v1
kind: Service
metadata:
  name: x-tunnel-server-svc
  namespace: kube-system
  labels:
    name: tunnel-server
spec:
  type: NodePort 
  ports:
  - port: 10263
    targetPort: 10263
    name: https
  - port: 10262
    targetPort: 10262
    nodePort: 31008
    name: tcp
  selector:
    k8s-app: tunnel-server
`
	TunnelServerInternalService = `
apiVersion: v1
kind: Service
metadata:
  name: x-tunnel-server-internal-svc
  namespace: kube-system
  labels:
    name: tunnel-server
spec:
  ports:
    - port: 10250
      targetPort: 10263
      name: https
    - port: 10255
      targetPort: 10264
      name: http
  selector:
    k8s-app: tunnel-server
`

	TunnelServerConfigMap = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-server-cfg
  namespace: kube-system
data:
  localhost-proxy-ports: "10266, 10267"
  http-proxy-ports: ""
  https-proxy-ports: ""
  dnat-ports-pair: ""
`
	TunnelServerDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tunnel-server
  namespace: kube-system
  labels:
    k8s-app: tunnel-server
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: tunnel-server
  template:
    metadata:
      labels:
        k8s-app: tunnel-server
    spec:
      hostNetwork: true
      serviceAccountName: tunnel-server
      restartPolicy: Always
      volumes:
      - name: tunnel-server-dir
        hostPath:
          path: /var/lib/tunnel-server
          type: DirectoryOrCreate
      tolerations:
      - operator: "Exists"
      nodeSelector:
        beta.kubernetes.io/arch: {{.arch}}
        beta.kubernetes.io/os: linux
        {{.edgeWorkerLabel}}: "false"
      containers:
      - name: tunnel-server
        image: {{.image}} 
        imagePullPolicy: IfNotPresent
        command:
        - tunnel-server
        args:
        - --bind-address=$(NODE_IP)
        - --insecure-bind-address=$(NODE_IP)
        - --server-count=1
          {{if .certIP }}
        - --cert-ips={{.certIP}}
          {{end}}
        env:
        - name: NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        securityContext:
          capabilities:
            add: ["NET_ADMIN", "NET_RAW"]
        volumeMounts:
        - name: tunnel-server-dir
          mountPath: /var/lib/tunnel-server
`
)
