package util

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

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	insecureListenAddr = "1.1.1.1:10264"
	secureListenAddr   = "1.1.1.1:10263"
)

func TestResolveProxyPortsAndMappings(t *testing.T) {
	testcases := map[string]struct {
		configMap    *corev1.ConfigMap
		expectResult struct {
			ports        []string
			portMappings map[string]string
			err          error
		}
	}{
		"setting ports on dnat-ports-pair": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"dnat-ports-pair": "9100=10264",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100"},
				portMappings: map[string]string{
					"9100": insecureListenAddr,
				},
				err: nil,
			},
		},
		"setting ports on http-proxy-ports": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"http-proxy-ports": "9100,9200",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100", "9200"},
				portMappings: map[string]string{
					"9100": insecureListenAddr,
					"9200": insecureListenAddr,
				},
				err: nil,
			},
		},
		"setting ports on https-proxy-ports": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"https-proxy-ports": "9100,9200",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100", "9200"},
				portMappings: map[string]string{
					"9100": secureListenAddr,
					"9200": secureListenAddr,
				},
				err: nil,
			},
		},
		"setting ports on http-proxy-ports and https-proxy-ports": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"http-proxy-ports":  "9100,9200",
					"https-proxy-ports": "9300,9400",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100", "9200", "9300", "9400"},
				portMappings: map[string]string{
					"9100": insecureListenAddr,
					"9200": insecureListenAddr,
					"9300": secureListenAddr,
					"9400": secureListenAddr,
				},
				err: nil,
			},
		},
	}

	for k, tt := range testcases {
		t.Run(k, func(t *testing.T) {
			ports, portMappings, err := resolveProxyPortsAndMappings(tt.configMap, insecureListenAddr, secureListenAddr)
			if tt.expectResult.err != err {
				t.Errorf("expect error: %v, but got error: %v", tt.expectResult.err, err)
			}

			// check the ports
			if len(tt.expectResult.ports) != len(ports) {
				t.Errorf("expect %d ports, but got %d ports", len(tt.expectResult.ports), len(ports))
			}

			foundPort := false
			for i := range tt.expectResult.ports {
				foundPort = false
				for j := range ports {
					if tt.expectResult.ports[i] == ports[j] {
						foundPort = true
						break
					}
				}

				if !foundPort {
					t.Errorf("expect %v ports, but got ports %v", tt.expectResult.ports, ports)
					break
				}
			}

			for i := range ports {
				foundPort = false
				for j := range tt.expectResult.ports {
					if tt.expectResult.ports[j] == ports[i] {
						foundPort = true
						break
					}
				}

				if !foundPort {
					t.Errorf("expect %v ports, but got ports %v", tt.expectResult.ports, ports)
					break
				}
			}

			// check the portMappings
			if len(tt.expectResult.portMappings) != len(portMappings) {
				t.Errorf("expect port mappings %v, but got port mappings %v", tt.expectResult.portMappings, portMappings)
			}

			for port, v := range tt.expectResult.portMappings {
				if gotV, ok := portMappings[port]; !ok {
					t.Errorf("expect port %s, but not got port", k)
				} else if v != gotV {
					t.Errorf("port(%s): expect dst value %s, but got dst value %s", k, v, gotV)
				}
				delete(portMappings, port)
			}
		})
	}
}

func TestRunMetaServer(t *testing.T) {
	var addr string = ":9090"
	RunMetaServer(addr)

	c := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				// When I use 127.0.0.0:9090, the unit test will occur 404 not found error but the page works healthily
				return net.Dial("tcp", "localhost:9090")
			},
		},
	}

	tests := []struct {
		desc string
		req  *http.Request
		code int
	}{
		{
			desc: "test metrics page",
			req: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "http",
					Host:   "localhost:9090",
					Path:   "/metrics",
				},
				Body: nil,
			},
			code: http.StatusOK,
		},
		{
			desc: "test pprof index page",
			req: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "http",
					Host:   "localhost:9090",
					Path:   "/debug/pprof",
				},
				Body: nil,
			},
			code: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			resp, err := c.Do(tt.req)
			if err != nil {
				t.Fatalf("fail to send request to the server: %v", err)
			}
			if resp.StatusCode != tt.code {
				t.Fatalf("the response status code is incorrect, expect: %d, get: %d", tt.code, resp.StatusCode)
			}
		})
	}

}
