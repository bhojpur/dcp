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
	"os/exec"
	"strings"
	"testing"

	testutil "github.com/bhojpur/dcp/tests/integration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var customEtcdArgsServer *testutil.DcpServer
var customEtcdArgsServerArgs = []string{
	"--cluster-init",
	"--etcd-arg quota-backend-bytes=858993459",
}
var testLock int

var _ = BeforeSuite(func() {
	if !testutil.IsExistingServer() {
		var err error
		testLock, err = testutil.DcpTestLock()
		Expect(err).ToNot(HaveOccurred())
		customEtcdArgsServer, err = testutil.DcpStartServer(customEtcdArgsServerArgs...)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = Describe("custom etcd args", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() && !testutil.ServerArgsPresent(customEtcdArgsServerArgs) {
			Skip("Test needs Bhojpur DCP server with: " + strings.Join(customEtcdArgsServerArgs, " "))
		}
	})
	When("a custom quota backend bytes is specified", func() {
		It("renders a config file with the correct entry", func() {
			Eventually(func() (string, error) {
				var cmd *exec.Cmd
				grepCmd := "grep"
				grepCmdArgs := []string{"quota-backend-bytes", "/var/lib/bhojpur/dcp/server/db/etcd/config"}
				if testutil.IsRoot() {
					cmd = exec.Command(grepCmd, grepCmdArgs...)
				} else {
					fullGrepCmd := append([]string{grepCmd}, grepCmdArgs...)
					cmd = exec.Command("sudo", fullGrepCmd...)
				}
				byteOut, err := cmd.CombinedOutput()
				return string(byteOut), err
			}, "45s", "5s").Should(MatchRegexp(".*quota-backend-bytes: 858993459.*"))
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() {
		Expect(testutil.DcpKillServer(customEtcdArgsServer)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, "")).To(Succeed())
	}
})

func Test_IntegrationCustomEtcdArgs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom etcd Arguments")
}
