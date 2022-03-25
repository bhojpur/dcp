package flannel

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
	"io/ioutil"
	"net"
	"regexp"
	"strings"
	"testing"

	"github.com/bhojpur/dcp/pkg/cloud/daemons/config"
)

func stringToCIDR(s string) []*net.IPNet {
	var netCidrs []*net.IPNet
	for _, v := range strings.Split(s, ",") {
		_, parsed, _ := net.ParseCIDR(v)
		netCidrs = append(netCidrs, parsed)
	}
	return netCidrs
}

func Test_findNetMode(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    int
		wantErr bool
	}{
		{"dual-stack", "10.42.0.0/16,2001:cafe:22::/56", ipv4 + ipv6, false},
		{"ipv4 only", "10.42.0.0/16", ipv4, false},
		{"ipv6 only", "2001:cafe:42:0::/56", ipv6, false},
		{"wrong input", "wrong", 0, true},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			netCidrs := stringToCIDR(tt.args)
			got, err := findNetMode(netCidrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("findNetMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findNetMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createFlannelConf(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		wantConfig []string
		wantErr    bool
	}{
		{"dual-stack", "10.42.0.0/16,2001:cafe:22::/56", []string{"\"Network\": \"10.42.0.0/16\"", "\"IPv6Network\": \"2001:cafe:22::/56\"", "\"EnableIPv6\": true"}, false},
		{"ipv4 only", "10.42.0.0/16", []string{"\"Network\": \"10.42.0.0/16\"", "\"IPv6Network\": \"::/0\"", "\"EnableIPv6\": false"}, false},
	}
	var containerd = config.Containerd{}
	for _, tt := range tests {
		var agent = config.Agent{}
		agent.ClusterCIDR = stringToCIDR(tt.args)[0]
		agent.ClusterCIDRs = stringToCIDR(tt.args)
		var nodeConfig = &config.Node{Docker: false, ContainerRuntimeEndpoint: "", NoFlannel: false, SELinux: false, FlannelBackend: "vxlan", FlannelConfFile: "test_file", FlannelConfOverride: false, FlannelIface: nil, Containerd: containerd, Images: "", AgentConfig: agent, Token: "", Certificate: nil, ServerHTTPSPort: 0}

		t.Run(tt.name, func(t *testing.T) {
			if err := createFlannelConf(nodeConfig); (err != nil) != tt.wantErr {
				t.Errorf("createFlannelConf() error = %v, wantErr %v", err, tt.wantErr)
			}
			data, err := ioutil.ReadFile("test_file")
			if err != nil {
				t.Errorf("Something went wrong when reading the flannel config file")
			}
			for _, config := range tt.wantConfig {
				isExist, _ := regexp.Match(config, data)
				if !isExist {
					t.Errorf("Config is wrong, %s is not present", config)
				}
			}
		})
	}
}
