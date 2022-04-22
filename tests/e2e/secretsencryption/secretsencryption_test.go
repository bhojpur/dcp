package secretsencryption

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
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bhojpur/dcp/tests/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Valid nodeOS: generic/ubuntu2004, opensuse/Leap-15.3.x86_64, dweomer/microos.amd64
var nodeOS = flag.String("nodeOS", "generic/ubuntu2004", "VM operating system")
var serverCount = flag.Int("serverCount", 3, "number of server nodes")

// Environment Variables Info:
// E2E_RELEASE_VERSION=v1.23.1+dcp2 or nil for latest commit from master

func Test_E2EClusterValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	flag.Parse()
	RunSpecs(t, "Create Cluster Test Suite")
}

var (
	kubeConfigFile  string
	serverNodeNames []string
)

var _ = Describe("Verify Secrets Encryption Rotation", func() {
	Context("Cluster :", func() {
		It("Starts up with no issues", func() {
			var err error
			serverNodeNames, _, err = e2e.CreateCluster(*nodeOS, *serverCount, 0)
			Expect(err).NotTo(HaveOccurred(), e2e.GetVagrantLog)
			fmt.Println("CLUSTER CONFIG")
			fmt.Println("OS:", *nodeOS)
			fmt.Println("Server Nodes:", serverNodeNames)
			kubeConfigFile, err = e2e.GenKubeConfigFile(serverNodeNames[0])
			Expect(err).NotTo(HaveOccurred())
		})

		It("Checks node and pod status", func() {
			fmt.Printf("\nFetching node status\n")
			Eventually(func(g Gomega) {
				nodes, err := e2e.ParseNodes(kubeConfigFile, false)
				g.Expect(err).NotTo(HaveOccurred())
				for _, node := range nodes {
					g.Expect(node.Status).Should(Equal("Ready"))
				}
			}, "420s", "5s").Should(Succeed())
			_, _ = e2e.ParseNodes(kubeConfigFile, true)

			fmt.Printf("\nFetching pods status\n")
			Eventually(func(g Gomega) {
				pods, err := e2e.ParsePods(kubeConfigFile, false)
				g.Expect(err).NotTo(HaveOccurred())
				for _, pod := range pods {
					if strings.Contains(pod.Name, "helm-install") {
						g.Expect(pod.Status).Should(Equal("Completed"), pod.Name)
					} else {
						g.Expect(pod.Status).Should(Equal("Running"), pod.Name)
					}
				}
			}, "420s", "5s").Should(Succeed())
			_, _ = e2e.ParsePods(kubeConfigFile, true)
		})

		It("Deploys several secrets", func() {
			_, err := e2e.DeployWorkload("secrets.yaml", kubeConfigFile, false)
			Expect(err).NotTo(HaveOccurred(), "Secrets not deployed")
		})

		It("Verifies encryption start stage", func() {
			cmd := "sudo dcp secrets-encrypt status"
			for _, nodeName := range serverNodeNames {
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).Should(ContainSubstring("Encryption Status: Enabled"))
				Expect(res).Should(ContainSubstring("Current Rotation Stage: start"))
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: All hashes match"))
			}
		})

		It("Prepares for Secrets-Encryption Rotation", func() {
			cmd := "sudo dcp secrets-encrypt prepare"
			res, err := e2e.RunCmdOnNode(cmd, serverNodeNames[0])
			Expect(err).NotTo(HaveOccurred(), res)
			for i, nodeName := range serverNodeNames {
				cmd := "sudo dcp secrets-encrypt status"
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred(), res)
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: hash does not match"))
				if i == 0 {
					Expect(res).Should(ContainSubstring("Current Rotation Stage: prepare"))
				} else {
					Expect(res).Should(ContainSubstring("Current Rotation Stage: start"))
				}
			}
		})

		It("Restarts Bhojpur DCP servers", func() {
			Expect(e2e.RestartCluster(serverNodeNames)).To(Succeed())
		})

		It("Checks node and pod status", func() {
			Eventually(func(g Gomega) {
				nodes, err := e2e.ParseNodes(kubeConfigFile, false)
				g.Expect(err).NotTo(HaveOccurred())
				for _, node := range nodes {
					g.Expect(node.Status).Should(Equal("Ready"))
				}
			}, "420s", "5s").Should(Succeed())

			Eventually(func(g Gomega) {
				pods, err := e2e.ParsePods(kubeConfigFile, false)
				g.Expect(err).NotTo(HaveOccurred())
				for _, pod := range pods {
					if strings.Contains(pod.Name, "helm-install") {
						g.Expect(pod.Status).Should(Equal("Completed"), pod.Name)
					} else {
						g.Expect(pod.Status).Should(Equal("Running"), pod.Name)
					}
				}
			}, "420s", "5s").Should(Succeed())
		})

		It("Verifies encryption prepare stage", func() {
			cmd := "sudo dcp secrets-encrypt status"
			for _, nodeName := range serverNodeNames {
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).Should(ContainSubstring("Encryption Status: Enabled"))
				Expect(res).Should(ContainSubstring("Current Rotation Stage: prepare"))
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: All hashes match"))
			}
		})

		It("Rotates the Secrets-Encryption Keys", func() {
			cmd := "sudo dcp secrets-encrypt rotate"
			res, err := e2e.RunCmdOnNode(cmd, serverNodeNames[0])
			Expect(err).NotTo(HaveOccurred(), res)
			for i, nodeName := range serverNodeNames {
				cmd := "sudo dcp secrets-encrypt status"
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred(), res)
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: hash does not match"))
				if i == 0 {
					Expect(res).Should(ContainSubstring("Current Rotation Stage: rotate"))
				} else {
					Expect(res).Should(ContainSubstring("Current Rotation Stage: prepare"))
				}
			}
		})

		It("Restarts Bhojpur DCP servers", func() {
			Expect(e2e.RestartCluster(serverNodeNames)).To(Succeed())
			time.Sleep(20 * time.Second)
		})

		It("Verifies encryption rotate stage", func() {
			cmd := "sudo dcp secrets-encrypt status"
			for _, nodeName := range serverNodeNames {
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).Should(ContainSubstring("Encryption Status: Enabled"))
				Expect(res).Should(ContainSubstring("Current Rotation Stage: rotate"))
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: All hashes match"))
			}
		})

		It("Reencrypts the Secrets-Encryption Keys", func() {
			cmd := "sudo dcp secrets-encrypt reencrypt"
			res, err := e2e.RunCmdOnNode(cmd, serverNodeNames[0])
			Expect(err).NotTo(HaveOccurred(), res)

			cmd = "sudo dcp secrets-encrypt status"
			Eventually(func() (string, error) {
				return e2e.RunCmdOnNode(cmd, serverNodeNames[0])
			}, "30s", "5s").Should(ContainSubstring("Current Rotation Stage: reencrypt_finished"))

			for _, nodeName := range serverNodeNames[1:] {
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred(), res)
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: hash does not match"))
				Expect(res).Should(ContainSubstring("Current Rotation Stage: rotate"))
			}
		})

		It("Restarts Bhojpur DCP Servers", func() {
			Expect(e2e.RestartCluster(serverNodeNames)).To(Succeed())
			time.Sleep(20 * time.Second)
		})

		It("Verifies Encryption Reencrypt Stage", func() {
			cmd := "sudo dcp secrets-encrypt status"
			for _, nodeName := range serverNodeNames {
				res, err := e2e.RunCmdOnNode(cmd, nodeName)
				Expect(err).NotTo(HaveOccurred())
				Expect(res).Should(ContainSubstring("Encryption Status: Enabled"))
				Expect(res).Should(ContainSubstring("Current Rotation Stage: reencrypt_finished"))
				Expect(res).Should(ContainSubstring("Server Encryption Hashes: All hashes match"))
			}
		})
	})

	It("Disables encryption", func() {
		cmd := "sudo dcp secrets-encrypt disable"
		res, err := e2e.RunCmdOnNode(cmd, serverNodeNames[0])
		Expect(err).NotTo(HaveOccurred(), res)

		cmd = "sudo dcp secrets-encrypt reencrypt -f --skip"
		res, err = e2e.RunCmdOnNode(cmd, serverNodeNames[0])
		Expect(err).NotTo(HaveOccurred(), res)

		cmd = "sudo dcp secrets-encrypt status"
		Eventually(func() (string, error) {
			return e2e.RunCmdOnNode(cmd, serverNodeNames[0])
		}, "30s", "5s").Should(ContainSubstring("Current Rotation Stage: reencrypt_finished"))

		for i, nodeName := range serverNodeNames {
			res, err := e2e.RunCmdOnNode(cmd, nodeName)
			Expect(err).NotTo(HaveOccurred(), res)
			if i == 0 {
				Expect(res).Should(ContainSubstring("Encryption Status: Disabled"))
			} else {
				Expect(res).Should(ContainSubstring("Encryption Status: Enabled"))
			}
		}
	})

	It("Restarts Bhojpur DCP servers", func() {
		Expect(e2e.RestartCluster(serverNodeNames)).To(Succeed())
		time.Sleep(20 * time.Second)
	})

	It("Verifies encryption disabled on all nodes", func() {
		cmd := "sudo dcp secrets-encrypt status"
		for _, nodeName := range serverNodeNames {
			Expect(e2e.RunCmdOnNode(cmd, nodeName)).Should(ContainSubstring("Encryption Status: Disabled"))
		}
	})

	It("Enables encryption", func() {
		cmd := "sudo dcp secrets-encrypt enable"
		res, err := e2e.RunCmdOnNode(cmd, serverNodeNames[0])
		Expect(err).NotTo(HaveOccurred(), res)

		cmd = "sudo dcp secrets-encrypt reencrypt -f --skip"
		res, err = e2e.RunCmdOnNode(cmd, serverNodeNames[0])
		Expect(err).NotTo(HaveOccurred(), res)

		cmd = "sudo dcp secrets-encrypt status"
		Eventually(func() (string, error) {
			return e2e.RunCmdOnNode(cmd, serverNodeNames[0])
		}, "30s", "5s").Should(ContainSubstring("Current Rotation Stage: reencrypt_finished"))
	})

	It("Restarts Bhojpur DCP servers", func() {
		Expect(e2e.RestartCluster(serverNodeNames)).To(Succeed())
		time.Sleep(20 * time.Second)
	})

	It("Verifies encryption enabled on all nodes", func() {
		cmd := "sudo dcp secrets-encrypt status"
		for _, nodeName := range serverNodeNames {
			Expect(e2e.RunCmdOnNode(cmd, nodeName)).Should(ContainSubstring("Encryption Status: Enabled"))
		}
	})

})

var failed = false
var _ = AfterEach(func() {
	failed = failed || CurrentSpecReport().Failed()
})

var _ = AfterSuite(func() {
	if failed {
		fmt.Println("FAILED!")
	} else {
		Expect(e2e.DestroyCluster()).To(Succeed())
		Expect(os.Remove(kubeConfigFile)).To(Succeed())
	}
})