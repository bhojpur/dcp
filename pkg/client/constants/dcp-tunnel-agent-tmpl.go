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
	TunnelAgentDaemonSet = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: tunnel-agent
  name: tunnel-agent
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: tunnel-agent
  template:
    metadata:
      labels:
        k8s-app: tunnel-agent
    spec:
      nodeSelector:
        beta.kubernetes.io/os: linux
        {{.edgeWorkerLabel}}: "true"
      containers:
      - command:
        - tunnel-agent
        args:
        - --node-name=$(NODE_NAME)
          {{if  .tunnelServerAddress }}
        - --tunnelserver-addr={{.tunnelServerAddress}}
          {{end}}
        image: {{.image}}
        imagePullPolicy: IfNotPresent
        name: tunnel-agent
        volumeMounts:
        - name: k8s-dir
          mountPath: /etc/kubernetes
        - name: kubelet-pki
          mountPath: /var/lib/kubelet/pki
        - name: tunnel-agent-dir
          mountPath: /var/lib/tunnel-agent
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: NODE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
      hostNetwork: true
      restartPolicy: Always
      volumes:
      - name: k8s-dir
        hostPath:
          path: /etc/kubernetes
          type: Directory
      - name: kubelet-pki
        hostPath:
          path: /var/lib/kubelet/pki
          type: Directory
      - name: tunnel-agent-dir
        hostPath:
          path: /var/lib/tunnel-agent
          type: DirectoryOrCreate
`
)
