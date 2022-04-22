package snapshot_test

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
	"regexp"
	"strings"
	"testing"
	"time"

	testutil "github.com/bhojpur/dcp/tests/integration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var server *testutil.DcpServer
var serverArgs = []string{"--cluster-init"}
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

var _ = Describe("etcd snapshots", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() && !testutil.ServerArgsPresent(serverArgs) {
			Skip("Test needs Bhojpur DCP server with: " + strings.Join(serverArgs, " "))
		}
	})
	When("a new etcd is created", func() {
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl", "get pods -A")
			}, "180s", "5s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("saves an etcd snapshot", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "save")).
				To(ContainSubstring("saved"))
		})
		It("list snapshots", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "ls")).
				To(MatchRegexp(`:///var/lib/bhojpur/dcp/server/db/snapshots/on-demand`))
		})
		It("deletes a snapshot", func() {
			lsResult, err := testutil.DcpCmd("etcd-snapshot", "ls")
			Expect(err).ToNot(HaveOccurred())
			reg, err := regexp.Compile(`on-demand[^\s]+`)
			Expect(err).ToNot(HaveOccurred())
			snapshotName := reg.FindString(lsResult)
			Expect(testutil.DcpCmd("etcd-snapshot", "delete", snapshotName)).
				To(ContainSubstring("Removing the given locally stored etcd snapshot"))
		})
	})
	When("saving a custom name", func() {
		It("saves an etcd snapshot with a custom name", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "save --name ALIVEBEEF")).
				To(ContainSubstring("Saving etcd snapshot to /var/lib/bhojpur/dcp/server/db/snapshots/ALIVEBEEF"))
		})
		It("deletes that snapshot", func() {
			lsResult, err := testutil.DcpCmd("etcd-snapshot", "ls")
			Expect(err).ToNot(HaveOccurred())
			reg, err := regexp.Compile(`ALIVEBEEF[^\s]+`)
			Expect(err).ToNot(HaveOccurred())
			snapshotName := reg.FindString(lsResult)
			Expect(testutil.DcpCmd("etcd-snapshot", "delete", snapshotName)).
				To(ContainSubstring("Removing the given locally stored etcd snapshot"))
		})
	})
	When("using etcd snapshot prune", func() {
		It("saves 3 different snapshots", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "save -name PRUNE_TEST")).
				To(ContainSubstring("saved"))
			time.Sleep(1 * time.Second)
			Expect(testutil.DcpCmd("etcd-snapshot", "save -name PRUNE_TEST")).
				To(ContainSubstring("saved"))
			time.Sleep(1 * time.Second)
			Expect(testutil.DcpCmd("etcd-snapshot", "save -name PRUNE_TEST")).
				To(ContainSubstring("saved"))
			time.Sleep(1 * time.Second)
		})
		It("lists all 3 snapshots", func() {
			lsResult, err := testutil.DcpCmd("etcd-snapshot", "ls")
			Expect(err).ToNot(HaveOccurred())
			reg, err := regexp.Compile(`:///var/lib/bhojpur/dcp/server/db/snapshots/PRUNE_TEST`)
			Expect(err).ToNot(HaveOccurred())
			sepLines := reg.FindAllString(lsResult, -1)
			Expect(sepLines).To(HaveLen(3))
		})
		It("prunes snapshots down to 2", func() {
			Expect(testutil.DcpCmd("etcd-snapshot", "prune --snapshot-retention 2 --name PRUNE_TEST")).
				To(ContainSubstring("Removing local snapshot"))
			lsResult, err := testutil.DcpCmd("etcd-snapshot", "ls")
			Expect(err).ToNot(HaveOccurred())
			reg, err := regexp.Compile(`:///var/lib/bhojpur/dcp/server/db/snapshots/PRUNE_TEST`)
			Expect(err).ToNot(HaveOccurred())
			sepLines := reg.FindAllString(lsResult, -1)
			Expect(sepLines).To(HaveLen(2))
		})
		It("cleans up remaining snapshots", func() {
			lsResult, err := testutil.DcpCmd("etcd-snapshot", "ls")
			Expect(err).ToNot(HaveOccurred())
			reg, err := regexp.Compile(`PRUNE_TEST[^\s]+`)
			Expect(err).ToNot(HaveOccurred())
			for _, snapshotName := range reg.FindAllString(lsResult, -1) {
				Expect(testutil.DcpCmd("etcd-snapshot", "delete", snapshotName)).
					To(ContainSubstring("Removing the given locally stored etcd snapshot"))
			}
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() {
		Expect(testutil.DcpKillServer(server)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, "")).To(Succeed())
	}
})

func Test_IntegrationEtcdSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcd Snapshot Suite")
}
