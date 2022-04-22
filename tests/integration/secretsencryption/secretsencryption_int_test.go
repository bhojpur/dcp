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
	"regexp"
	"testing"
	"time"

	testutil "github.com/bhojpur/dcp/tests/integration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var secretsEncryptionServer *testutil.DcpServer
var secretsEncryptionDataDir = "/tmp/dcpse"
var secretsEncryptionServerArgs = []string{"--secrets-encryption", "-d", secretsEncryptionDataDir}
var testLock int

var _ = BeforeSuite(func() {
	if !testutil.IsExistingServer() {
		var err error
		testLock, err = testutil.DcpTestLock()
		Expect(err).ToNot(HaveOccurred())
		secretsEncryptionServer, err = testutil.DcpStartServer(secretsEncryptionServerArgs...)
		Expect(err).ToNot(HaveOccurred())
	}
})

var _ = Describe("secrets encryption rotation", func() {
	BeforeEach(func() {
		if testutil.IsExistingServer() {
			Skip("Test does not support running on existing Bhojpur DCP servers")
		}
	})
	When("A server starts with secrets encryption", func() {
		It("starts up with no problems", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get pods -A")
			}, "180s", "1s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("it creates a encryption key", func() {
			result, err := testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("Encryption Status: Enabled"))
			Expect(result).To(ContainSubstring("Current Rotation Stage: start"))
		})
	})
	When("A server rotates encryption keys", func() {
		It("it prepares to rotate", func() {
			Expect(testutil.DcpCmd("secrets-encrypt prepare -d", secretsEncryptionDataDir)).
				To(ContainSubstring("prepare completed successfully"))

			result, err := testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("Current Rotation Stage: prepare"))
			reg, err := regexp.Compile(`AES-CBC.+aescbckey.*`)
			Expect(err).ToNot(HaveOccurred())
			keys := reg.FindAllString(result, -1)
			Expect(keys).To(HaveLen(2))
			Expect(keys[0]).To(ContainSubstring("aescbckey"))
			Expect(keys[1]).To(ContainSubstring("aescbckey-" + fmt.Sprint(time.Now().Year())))
		})
		It("restarts the server", func() {
			var err error
			Expect(testutil.DcpKillServer(secretsEncryptionServer)).To(Succeed())
			secretsEncryptionServer, err = testutil.DcpStartServer(secretsEncryptionServerArgs...)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get pods -A")
			}, "180s", "1s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
		})
		It("rotates the keys", func() {
			Eventually(func() (string, error) {
				return testutil.DcpCmd("secrets-encrypt rotate -d", secretsEncryptionDataDir)
			}, "10s", "2s").Should(ContainSubstring("rotate completed successfully"))

			result, err := testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("Current Rotation Stage: rotate"))
			reg, err := regexp.Compile(`AES-CBC.+aescbckey.*`)
			Expect(err).ToNot(HaveOccurred())
			keys := reg.FindAllString(result, -1)
			Expect(keys).To(HaveLen(2))
			Expect(keys[0]).To(ContainSubstring("aescbckey-" + fmt.Sprint(time.Now().Year())))
			Expect(keys[1]).To(ContainSubstring("aescbckey"))
		})
		It("restarts the server", func() {
			var err error
			Expect(testutil.DcpKillServer(secretsEncryptionServer)).To(Succeed())
			secretsEncryptionServer, err = testutil.DcpStartServer(secretsEncryptionServerArgs...)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get pods -A")
			}, "180s", "1s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
			time.Sleep(10 * time.Second)
		})
		It("reencrypts the keys", func() {
			Expect(testutil.DcpCmd("secrets-encrypt reencrypt -d", secretsEncryptionDataDir)).
				To(ContainSubstring("reencryption started"))
			Eventually(func() (string, error) {
				return testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			}, "45s", "2s").Should(ContainSubstring("Current Rotation Stage: reencrypt_finished"))
			result, err := testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			reg, err := regexp.Compile(`AES-CBC.+aescbckey.*`)
			Expect(err).ToNot(HaveOccurred())
			keys := reg.FindAllString(result, -1)
			Expect(keys).To(HaveLen(1))
			Expect(keys[0]).To(ContainSubstring("aescbckey-" + fmt.Sprint(time.Now().Year())))
		})
	})
	When("A server disables encryption", func() {
		It("it triggers the disable", func() {
			Expect(testutil.DcpCmd("secrets-encrypt disable -d", secretsEncryptionDataDir)).
				To(ContainSubstring("secrets-encryption disabled"))

			result, err := testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("Encryption Status: Disabled"))
		})
		It("restarts the server", func() {
			var err error
			Expect(testutil.DcpKillServer(secretsEncryptionServer)).To(Succeed())
			secretsEncryptionServer, err = testutil.DcpStartServer(secretsEncryptionServerArgs...)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() (string, error) {
				return testutil.DcpCmd("kubectl get pods -A")
			}, "180s", "1s").Should(MatchRegexp("kube-system.+coredns.+1\\/1.+Running"))
			time.Sleep(10 * time.Second)
		})
		It("reencrypts the keys", func() {
			result, err := testutil.DcpCmd("secrets-encrypt reencrypt -f --skip -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("reencryption started"))

			result, err = testutil.DcpCmd("secrets-encrypt status -d", secretsEncryptionDataDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring("Encryption Status: Disabled"))
		})
	})
})

var _ = AfterSuite(func() {
	if !testutil.IsExistingServer() {
		Expect(testutil.DcpKillServer(secretsEncryptionServer)).To(Succeed())
		Expect(testutil.DcpCleanup(testLock, secretsEncryptionDataDir)).To(Succeed())
	}
})

func Test_IntegrationSecretsEncryption(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Secrets Encryption Suite")
}
