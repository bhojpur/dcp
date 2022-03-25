package edgenode

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
	KubeletSvcPath      = "/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf"
	BhojpurDir          = "/var/lib/bhojpur"
	StaticPodPath       = "/etc/kubernetes/manifests"
	KubeCondfigPath     = "/etc/kubernetes/kubelet.conf"
	KubeCaFile          = "/etc/kubernetes/pki/ca.crt"
	EngineYamlName      = "dcpsvr.yaml"
	EngineComponentName = "dcpsvr"
	EngineNamespace     = "kube-system"
	EngineCmName        = "dcpsvr-cfg"
	KubeletConfName     = "kubelet.conf"
	KubeletSvcBackup    = "%s.bk"

	Hostname               = "/etc/hostname"
	KubeletHostname        = "--hostname-override=[^\"\\s]*"
	KubeletEnvironmentFile = "EnvironmentFile=.*"

	DaemonReload      = "systemctl daemon-reload"
	RestartKubeletSvc = "systemctl restart kubelet"

	ServerHealthzServer  = "127.0.0.1:10267"
	ServerHealthzURLPath = "/v1/healthz"
	DcpKubeletConf       = `
apiVersion: v1
clusters:
- cluster:
    server: http://127.0.0.1:10261
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    namespace: default
    user: default-auth
  name: default-context
current-context: default-context
kind: Config
preferences: {}
`
	EngineTemplate = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    k8s-app: dcpsvr
  name: dcpsvr
  namespace: kube-system
spec:
  volumes:
  - name: hub-dir
    hostPath:
      path: /var/lib/dcpsvr
      type: DirectoryOrCreate
  - name: kubernetes
    hostPath:
      path: /etc/kubernetes
      type: Directory
  - name: pem-dir
    hostPath:
      path: /var/lib/kubelet/pki
      type: Directory
  containers:
  - name: dcpsvr
    image: {{.image}}
    imagePullPolicy: IfNotPresent
    volumeMounts:
    - name: hub-dir
      mountPath: /var/lib/dcpsvr
    - name: kubernetes
      mountPath: /etc/kubernetes
    - name: pem-dir
      mountPath: /var/lib/kubelet/pki
    command:
    - dcpsvr
    - --v=2
    - --server-addr={{.kubernetesServerAddr}}
    - --node-name=$(NODE_NAME)
    - --join-token={{.joinToken}}
    - --working-mode={{.workingMode}}
      {{if .organizations }}
    - --hub-cert-organizations={{.organizations}}
      {{end}}
    livenessProbe:
      httpGet:
        host: 127.0.0.1
        path: /v1/healthz
        port: 10267
      initialDelaySeconds: 300
      periodSeconds: 5
      failureThreshold: 3
    resources:
      requests:
        cpu: 150m
        memory: 150Mi
      limits:
        memory: 300Mi
    securityContext:
      capabilities:
        add: ["NET_ADMIN", "NET_RAW"]
    env:
    - name: NODE_NAME
      valueFrom:
        fieldRef:
          fieldPath: spec.nodeName
  hostNetwork: true
  priorityClassName: system-node-critical
  priority: 2000001000
`
	EngineClusterRole = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: dcpsvr
rules:
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - get
  - apiGroups:
      - apps.bhojpur.net
    resources:
      - nodepools
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - configmaps
    resourceNames:
      - dcpsvr-cfg
    verbs:
      - list
      - watch
`
	EngineClusterRoleBinding = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: dcpsvr
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: dcpsvr
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: Group
    name: system:nodes
`
	EngineConfigMap = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: dcp-svr-cfg
  namespace: kube-system
data:
  cache_agents: ""`
)
