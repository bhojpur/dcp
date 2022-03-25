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
	TunnelServerAgentPort           = "10262"
	TunnelServerMasterPort          = "10263"
	TunnelServerMasterInsecurePort  = "10264"
	TunnelServerMetaPort            = "10265"
	TunnelAgentMetaPort             = "10266"
	TunnelServerServiceNs           = "kube-system"
	TunnelServerInternalServiceName = "x-tunnel-server-internal-svc"
	TunnelServerServiceName         = "x-tunnel-server-svc"
	TunnelServerAgentPortName       = "tcp"
	TunnelServerExternalAddrKey     = "x-tunnel-server-external-addr"
	TunnelEndpointsNs               = "kube-system"
	TunnelEndpointsName             = "x-tunnel-server-svc"
	TunnelDNSRecordConfigMapNs      = "kube-system"
	TunnelDNSRecordConfigMapName    = "%s-tunnel-nodes"
	TunnelDNSRecordNodeDataKey      = "tunnel-nodes"

	// Tunnel PKI related constants
	TunnelCSROrg                 = "bhojpur:tunnel"
	TunnelAgentCSRCN             = "tunnel-agent"
	TunneServerCSROrg            = "system:masters"
	TunneServerCSRCN             = "kube-apiserver-kubelet-client"
	TunnelCAFile                 = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	TunnelTokenFile              = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	TunnelServerCertDir          = "/var/lib/%s/pki"
	TunnelAgentCertDir           = "/var/lib/%s/pki"
	TunnelCSRApproverThreadiness = 2

	// name of the environment variables used in pod
	TunnelAgentPodIPEnv = "POD_IP"

	// name of the environment for selecting backend agent used in tunnel-server
	NodeIPKeyIndex     = "status.internalIP"
	ProxyHostHeaderKey = "X-Tunnel-Proxy-Host"
	ProxyDestHeaderKey = "X-Tunnel-Proxy-Dest"

	// The timeout seconds of reading a complete request from the apiserver
	TunnelANPInterceptorReadTimeoutSec = 10
	// The period between two keep-alive probes
	TunnelANPInterceptorKeepAlivePeriodSec = 10
	// The timeout seconds for the interceptor to proceed a complete read from the proxier
	TunnelANPProxierReadTimeoutSec = 10
	// probe the client every 10 seconds to ensure the connection is still active
	TunnelANPGrpcKeepAliveTimeSec = 10
	// wait 5 seconds for the probe ack before cutting the connection
	TunnelANPGrpcKeepAliveTimeoutSec = 5
)
