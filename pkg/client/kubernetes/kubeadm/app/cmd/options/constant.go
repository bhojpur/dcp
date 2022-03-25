package options

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
	// APIServerAdvertiseAddress flag sets the IP address the API Server will advertise it's listening on. Specify '0.0.0.0' to use the address of the default network interface.
	APIServerAdvertiseAddress = "apiserver-advertise-address"

	// APIServerBindPort flag sets the port for the API Server to bind to.
	APIServerBindPort = "apiserver-bind-port"

	// CertificatesDir flag sets the path where to save and read the certificates.
	CertificatesDir = "cert-dir"

	// CfgPath flag sets the path to kubeadm config file.
	CfgPath = "config"

	// DryRun flag instruct kubeadm to don't apply any changes; just output what would be done.
	DryRun = "dry-run"

	// IgnorePreflightErrors sets the path a list of checks whose errors will be shown as warnings. Example: 'IsPrivilegedUser,Swap'. Value 'all' ignores errors from all checks.
	IgnorePreflightErrors = "ignore-preflight-errors"

	// KubeconfigPath flag sets the kubeconfig file to use when talking to the cluster. If the flag is not set, a set of standard locations are searched for an existing KubeConfig file.
	KubeconfigPath = "kubeconfig"

	// KubernetesVersion flag sets the Kubernetes version for the control plane.
	KubernetesVersion = "kubernetes-version"

	// NodeCRISocket flag sets the CRI socket to connect to.
	NodeCRISocket = "cri-socket"

	// NodeName flag sets the node name.
	NodeName = "node-name"

	// TokenStr flags sets both the discovery-token and the tls-bootstrap-token when those values are not provided
	TokenStr = "token"

	// TokenDiscoveryCAHash flag instruct kubeadm to validate that the root CA public key matches this hash (for token-based discovery)
	TokenDiscoveryCAHash = "discovery-token-ca-cert-hash"

	// TokenDiscoverySkipCAHash flag instruct kubeadm to skip CA hash verification (for token-based discovery)
	TokenDiscoverySkipCAHash = "discovery-token-unsafe-skip-ca-verification"

	// ForceReset flag instruct kubeadm to reset the node without prompting for confirmation
	ForceReset = "force"

	// NodeType flag sets the type of worker node to edge or cloud.
	NodeType = "node-type"

	// Organizations flag sets the extra organizations of hub agent client certificate.
	Organizations = "organizations"

	// NodeLabels flag sets the labels for worker node.
	NodeLabels = "node-labels"

	// PauseImage flag sets the pause image for worker node.
	PauseImage = "pause-image"

	// EngineImage flag sets the Bhojpur DCP server engine image for worker node.
	EngineImage = "dcpsvr-image"

	// KubernetesResourceServer flag sets the address for download k8s node resources.
	KubernetesResourceServer = "kubernetes-resource-server"
)
