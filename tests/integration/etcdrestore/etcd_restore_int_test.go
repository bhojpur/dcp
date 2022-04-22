package restore_test

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

var server1, server2 *testutil.DcpServer
var tmpdDataDir = "/tmp/restoredatadir"
var clientCACertHash string
var testLock int
var restoreServerArgs = []string{"--cluster-init", "-t", "test", "-d", tmpdDataDir}
var _ = BeforeSuite(func() {
	if !testutil.IsExistingServer() {
		var err error
		testLock, err = testutil.DcpTestLock()
		Expect(err).ToNot(HaveOccurred())
		server1, err = testutil.DcpStartServer(restoreServerArgs...)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = Describe("etcd snapshot restore", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() && !testutil.ServerArgsPresent(restoreServerArgs) {
			Skip("Test needs Bhojpur DCP server with: " + strings.Join(restoreServerArgs, " "))
		}
	})
	When("a snapshot is restored on existing node", func() {
		It("etcd starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get", "pods", "-A")
			}, "360s", "5s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("create a workload", func() {
			result, err := testutil.DcpCmd("kubectl", "create", "-f", "./testdata/temp_depl.yaml")
			Expect(result).To(ContainSubstring("deployment.apps/nginx-deployment created"))
			Expect(err).NotTo(HaveOccurred())
		})
		It("saves an etcd snapshot", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "save", "-d", tmpdDataDir, "--name", "snapshot-to-restore")).
				To(ContainSubstring("saved"))
		})
		It("list snapshots", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "ls", "-d", tmpdDataDir)).
				To(MatchRegexp(`://` + tmpdDataDir + `/server/db/snapshots/snapshot-to-restore`))
		})
		// create another workload
		It("create a workload 2", func() {
			result, err := testutil.DcpCmd("kubectl", "create", "-f", "./testdata/temp_depl2.yaml")
			Expect(result).To(ContainSubstring("deployment.apps/nginx-deployment-post-snapshot created"))
			Expect(err).NotTo(HaveOccurred())
		})
		It("get Client CA cert hash", func() {
			// get md5sum of the CA certs
			var err error
			clientCACertHash, err = testutil.RunCommand("md5sum " + tmpdDataDir + "/server/tls/client-ca.crt | cut -f 1 -d' '")
			Expect(err).ToNot(HaveOccurred())
		})
		It("stop dcp", func() {
			Expect(testutil.DcpKillServer(server1)).To(Succeed())
		})
		It("restore the snapshot", func() {
			// get snapshot file
			filePath, err := testutil.RunCommand(`sudo find ` + tmpdDataDir + `/server -name "*snapshot-to-restore*"`)
			Expect(err).ToNot(HaveOccurred())
			filePath = strings.TrimSuffix(filePath, "\n")
			Eventually(func() (string, error) {
				return testutil.DcpCmd("server", "-d", tmpdDataDir, "--cluster-reset", "--token", "test", "--cluster-reset-restore-path", filePath)
			}, "360s", "5s").Should(ContainSubstring(`Etcd is running, restart without --cluster-reset flag now`))
		})
		It("start dcp server", func() {
			var err error
			server2, err = testutil.DcpStartServer(restoreServerArgs...)
			Expect(err).ToNot(HaveOccurred())
		})
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get", "pods", "-A")
			}, "360s", "5s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("Make sure Workload 1 exists", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get", "deployment", "nginx-deployment")
			}, "360s", "5s").Should(ContainSubstring("3/3"))
		})
		It("Make sure Workload 2 does not exists", func() {
			res, err := testutil.DcpCmd("kubectl", "get", "deployment", "nginx-deployment-post-snapshot")
			Expect(err).To(HaveOccurred())
			Expect(res).To(ContainSubstring("not found"))
		})
		It("check if CA cert hash matches", func() {
			// get md5sum of the CA certs
			var err error
			clientCACertHash2, err := testutil.RunCommand("md5sum " + tmpdDataDir + "/server/tls/client-ca.crt | cut -f 1 -d' '")
			Expect(err).ToNot(HaveOccurred())
			Expect(clientCACertHash2).To(Equal(clientCACertHash))
		})
		It("stop dcp", func() {
			Expect(testutil.DcpKillServer(server2)).To(Succeed())
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() {
		Expect(testutil.DcpKillServer(server1)).To(Succeed())
		Expect(testutil.DcpKillServer(server2)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, tmpdDataDir)).To(Succeed())
	}
})

func Test_IntegrationEtcdRestoreSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcd Restore Suite")
}
