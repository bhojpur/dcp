package integration

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
	"os"
	"strings"
	"testing"

	testutil "github.com/bhojpur/dcp/tests/integration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var dualStackServer *testutil.DcpServer
var dualStackServerArgs = []string{
	"--cluster-init",
	"--cluster-cidr 10.42.0.0/16,2001:cafe:42:0::/56",
	"--service-cidr 10.43.0.0/16,2001:cafe:42:1::/112",
	"--disable-network-policy",
}
var testLock int

var _ = BeforeSuite(func() {
	if !testutil.IsExistingServer() && os.Getenv("CI") != "true" {
		var err error
		testLock, err = testutil.DcpTestLock()
		Expect(err).ToNot(HaveOccurred())
		dualStackServer, err = testutil.DcpStartServer(dualStackServerArgs...)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = Describe("dual stack", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() && !testutil.ServerArgsPresent(dualStackServerArgs) {
			Skip("Test needs Bhojpur DCP server with: " + strings.Join(dualStackServerArgs, " "))
		} else if os.Getenv("CI") == "true" {
			Skip("Github environment does not support IPv6")
		}
	})
	When("a ipv4 and ipv6 cidr is present", func() {
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get", "pods", "-A")
			}, "180s", "5s").Should(MatchRegexp("kube-system.+traefik.+1\\/1.+Running"))
		})
		It("creates pods with two IPs", func() {
			podname, err := testutil.DcpCmd("kubectl", "get", "pods", "-n", "kube-system", "-o", "jsonpath={.items[?(@.metadata.labels.app\\.kubernetes\\.io/name==\"traefik\")].metadata.name}")
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "exec", podname, "-n", "kube-system", "--", "ip", "a")
			}, "5s", "1s").Should(ContainSubstring("2001:cafe:42:"))
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() && os.Getenv("CI") != "true" {
		Expect(testutil.DcpKillServer(dualStackServer)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, "")).To(Succeed())
	}
})

func Test_IntegrationDualStack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dual-Stack Suite")
}
