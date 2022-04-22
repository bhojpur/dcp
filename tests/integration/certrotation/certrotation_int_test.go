package cert_rotation_test

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
	"strings"
	"testing"

	testutil "github.com/bhojpur/dcp/tests/integration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const tmpdDataDir = "/tmp/certrotationtest"

var server, server2 *testutil.DcpServer
var serverArgs = []string{"--cluster-init", "-t", "test", "-d", tmpdDataDir}
var certHash, caCertHash string
var testLock int

var _ = BeforeSuite(func() {
	if !testutil.IsExistingServer() {
		var err error
		testLock, err = testutil.DcpTestLock()
		Expect(err).ToNot(HaveOccurred())
		server, err = testutil.DcpStartServer(serverArgs...)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = Describe("certificate rotation", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() && !testutil.ServerArgsPresent(serverArgs) {
			Skip("Test needs Bhojpur DCP server with: " + strings.Join(serverArgs, " "))
		}
	})
	When("a new server is created", func() {
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get pods -A")
			}, "180s", "5s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("get certificate hash", func() {
			// get md5sum of the CA certs
			var err error
			caCertHash, err = testutil.RunCommand("md5sum " + tmpdDataDir + "/server/tls/client-ca.crt | cut -f 1 -d' '")
			Expect(err).ToNot(HaveOccurred())
			certHash, err = testutil.RunCommand("md5sum " + tmpdDataDir + "/server/tls/serving-kube-apiserver.crt | cut -f 1 -d' '")
			Expect(err).ToNot(HaveOccurred())
		})
		It("stop dcp", func() {
			Expect(testutil.DcpKillServer(server)).To(Succeed())
		})
		It("certificate rotate", func() {
			_, err := testutil.DcpCmd("certificate", "rotate", "-d", tmpdDataDir)
			Expect(err).ToNot(HaveOccurred())

		})
		It("start dcp server", func() {
			var err error
			server2, err = testutil.DcpStartServer(serverArgs...)
			Expect(err).ToNot(HaveOccurred())
		})
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get", "pods", "-A")
			}, "360s", "5s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("get certificate hash", func() {
			// get md5sum of the CA certs
			var err error
			caCertHashAfter, err := testutil.RunCommand("md5sum " + tmpdDataDir + "/server/tls/client-ca.crt | cut -f 1 -d' '")
			Expect(err).ToNot(HaveOccurred())
			certHashAfter, err := testutil.RunCommand("md5sum " + tmpdDataDir + "/server/tls/serving-kube-apiserver.crt | cut -f 1 -d' '")
			Expect(err).ToNot(HaveOccurred())
			Expect(caCertHash).To(Not(Equal(certHashAfter)))
			Expect(caCertHash).To(Equal(caCertHashAfter))
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() {
		Expect(testutil.DcpKillServer(server)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, "")).To(Succeed())
		Expect(testutil.DcpKillServer(server2)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, tmpdDataDir)).To(Succeed())
	}
})

func Test_IntegrationCertRotation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cert rotation Suite")
}
