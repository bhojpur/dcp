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
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	testutil "github.com/bhojpur/dcp/tests/integration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var localStorageServer *testutil.DcpServer
var localStorageServerArgs = []string{"--cluster-init"}
var testLock int

var _ = BeforeSuite(func() {
	if !testutil.IsExistingServer() {
		var err error
		testLock, err = testutil.DcpTestLock()
		Expect(err).ToNot(HaveOccurred())
		localStorageServer, err = testutil.DcpStartServer(localStorageServerArgs...)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = Describe("local storage", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() && !testutil.ServerArgsPresent(localStorageServerArgs) {
			Skip("Test needs Bhojpur DCP server with: " + strings.Join(localStorageServerArgs, " "))
		}
	})
	When("a new local storage is created", func() {
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get pods -A")
			}, "90s", "1s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("creates a new pvc", func() {
			result, err := testutil.DcpCmd("kubectl create -f ./testdata/localstorage_pvc.yaml")
			Expect(result).To(ContainSubstring("persistentvolumeclaim/local-path-pvc created"))
			Expect(err).NotTo(HaveOccurred())
		})
		It("creates a new pod", func() {
			Expect(testutil.DcpCmd("kubectl create -f ./testdata/localstorage_pod.yaml")).
				To(ContainSubstring("pod/volume-test created"))
		})
		It("shows storage up in kubectl", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get --namespace=default pvc")
			}, "45s", "1s").Should(MatchRegexp(`local-path-pvc.+Bound`))
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get --namespace=default pv")
			}, "10s", "1s").Should(MatchRegexp(`pvc.+1Gi.+Bound`))
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get --namespace=default pod")
			}, "10s", "1s").Should(MatchRegexp(`volume-test.+Running`))
		})
		It("has proper folder permissions", func() {
			var dcpStorage = "/var/lib/bhojpur/dcp/storage"
			fileStat, err := os.Stat(dcpStorage)
			Expect(err).ToNot(HaveOccurred())
			Expect(fmt.Sprintf("%04o", fileStat.Mode().Perm())).To(Equal("0701"))

			pvResult, err := testutil.DcpCmd("kubectl get --namespace=default pv")
			Expect(err).ToNot(HaveOccurred())
			reg, err := regexp.Compile(`pvc[^\s]+`)
			Expect(err).ToNot(HaveOccurred())
			volumeName := reg.FindString(pvResult) + "_default_local-path-pvc"
			fileStat, err = os.Stat(dcpStorage + "/" + volumeName)
			Expect(err).ToNot(HaveOccurred())
			Expect(fmt.Sprintf("%04o", fileStat.Mode().Perm())).To(Equal("0777"))
		})
		It("deletes properly", func() {
			Expect(testutil.DcpCmd("kubectl delete --namespace=default --force pod volume-test")).
				To(ContainSubstring("pod \"volume-test\" force deleted"))
			Expect(testutil.DcpCmd("kubectl delete --namespace=default pvc local-path-pvc")).
				To(ContainSubstring("persistentvolumeclaim \"local-path-pvc\" deleted"))
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() {
		Expect(testutil.DcpKillServer(localStorageServer)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, "")).To(Succeed())
	}
})

func Test_IntegrationLocalStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Local Storage Suite")
}
