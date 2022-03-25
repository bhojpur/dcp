package dns

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
	"fmt"
	"net"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/bhojpur/dcp/pkg/projectinfo"
)

func isEdgeNode(node *corev1.Node) bool {
	isEdgeNode, ok := node.Labels[projectinfo.GetEdgeWorkerLabelKey()]
	if ok && isEdgeNode == "true" {
		return true
	}
	return false
}

func formatDNSRecord(ip, host string) string {
	return fmt.Sprintf("%s\t%s", ip, host)
}

// getNodeHostIP returns the provided node's "primary" IP
func getNodeHostIP(node *corev1.Node) (string, error) {
	// re-sort the addresses with InternalIPs first and then ExternalIPs
	allIPs := make([]net.IP, 0, len(node.Status.Addresses))
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ip := net.ParseIP(addr.Address)
			if ip != nil {
				allIPs = append(allIPs, ip)
				break
			}
		}
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			ip := net.ParseIP(addr.Address)
			if ip != nil {
				allIPs = append(allIPs, ip)
				break
			}
		}
	}
	if len(allIPs) == 0 {
		return "", fmt.Errorf("host IP unknown; known addresses: %v", node.Status.Addresses)
	}

	return allIPs[0].String(), nil
}

func removeRecordByHostname(records []string, hostname string) (result []string, changed bool) {
	result = make([]string, 0, len(records))
	for _, v := range records {
		if !strings.HasSuffix(v, hostname) {
			result = append(result, v)
		}
	}
	return result, len(records) != len(result)
}

func parseHostnameFromDNSRecord(record string) (string, error) {
	arr := strings.Split(record, "\t")
	if len(arr) != 2 {
		return "", fmt.Errorf("failed to parse hostname, invalid dns record %q", record)
	}
	return arr[1], nil
}

func addOrUpdateRecord(records []string, record string) (result []string, changed bool, err error) {
	hostname, err := parseHostnameFromDNSRecord(record)
	if err != nil {
		return nil, false, err
	}

	result = make([]string, len(records))
	copy(result, records)

	found := false
	for i, v := range result {
		if strings.HasSuffix(v, hostname) {
			found = true
			if v != record {
				result[i] = record
				changed = true
				break
			}
		}
	}

	if !found {
		result = append(result, record)
		changed = true
	}

	return result, changed, nil
}
