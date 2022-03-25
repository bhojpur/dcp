package phases

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
	"errors"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
	utilsexec "k8s.io/utils/exec"

	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	kubeutil "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/phases/kubelet"
	utilruntime "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/runtime"
)

// NewCleanupNodePhase creates a kubeadm workflow phase that cleanup the node
func NewCleanupNodePhase() workflow.Phase {
	return workflow.Phase{
		Name:    "cleanup-node",
		Aliases: []string{"cleanupnode"},
		Short:   "Run cleanup node.",
		Run:     runCleanupNode,
		InheritFlags: []string{
			options.NodeCRISocket,
		},
	}
}

func runCleanupNode(c workflow.RunData) error {
	r, ok := c.(resetData)
	if !ok {
		return errors.New("cleanup-node phase invoked with an invalid data struct")
	}

	// Try to stop the kubelet service
	klog.Infoln("[reset] Stopping the kubelet service")
	kubeutil.TryStopKubelet()

	// Try to unmount mounted directories under kubeadmconstants.KubeletRunDirectory in order to be able to remove the kubeadmconstants.KubeletRunDirectory directory later
	klog.Infof("[reset] Unmounting mounted directories in %q", kubeadmconstants.KubeletRunDirectory)
	// In case KubeletRunDirectory holds a symbolic link, evaluate it
	kubeletRunDir, err := absoluteKubeletRunDirectory()
	if err == nil {
		// Only clean absoluteKubeletRunDirectory if umountDirsCmd passed without error
		r.AddDirsToClean(kubeletRunDir)
	}

	klog.V(1).Info("[reset] Removing Kubernetes-managed containers")
	if err := removeContainers(utilsexec.New(), r.CRISocketPath()); err != nil {
		klog.Warningf("[reset] Failed to remove containers: %v", err)
	}

	r.AddDirsToClean("/var/lib/dockershim", "/var/run/kubernetes", "/var/lib/cni")

	// Remove contents from the config and pki directories
	klog.V(1).Infoln("[reset] Removing contents from the config and pki directories")
	certsDir := filepath.Join(kubeadmconstants.KubernetesDir, "pki")
	resetConfigDir(kubeadmconstants.KubernetesDir, certsDir)

	return nil
}

func absoluteKubeletRunDirectory() (string, error) {
	absoluteKubeletRunDirectory, err := filepath.EvalSymlinks(kubeadmconstants.KubeletRunDirectory)
	if err != nil {
		klog.Warningf("[reset] Failed to evaluate the %q directory. Skipping its unmount and cleanup: %v", kubeadmconstants.KubeletRunDirectory, err)
		return "", err
	}
	err = unmountKubeletDirectory(absoluteKubeletRunDirectory)
	if err != nil {
		klog.Warningf("[reset] Failed to unmount mounted directories in %s", kubeadmconstants.KubeletRunDirectory)
		return "", err
	}
	return absoluteKubeletRunDirectory, nil
}

func removeContainers(execer utilsexec.Interface, criSocketPath string) error {
	containerRuntime, err := utilruntime.NewContainerRuntime(execer, criSocketPath)
	if err != nil {
		return err
	}
	containers, err := containerRuntime.ListKubeContainers()
	if err != nil {
		return err
	}
	return containerRuntime.RemoveContainers(containers)
}

// resetConfigDir is used to cleanup the files kubeadm writes in /etc/kubernetes/.
func resetConfigDir(configPathDir, pkiPathDir string) {
	dirsToClean := []string{
		filepath.Join(configPathDir, kubeadmconstants.ManifestsSubDirName),
		pkiPathDir,
	}
	klog.Infof("[reset] Deleting contents of config directories: %v", dirsToClean)
	for _, dir := range dirsToClean {
		if err := CleanDir(dir); err != nil {
			klog.Warningf("[reset] Failed to delete contents of %q directory: %v", dir, err)
		}
	}

	filesToClean := []string{
		filepath.Join(configPathDir, kubeadmconstants.KubeletKubeConfigFileName),
	}
	klog.Infof("[reset] Deleting files: %v", filesToClean)
	for _, path := range filesToClean {
		if err := os.RemoveAll(path); err != nil {
			klog.Warningf("[reset] Failed to remove file: %q [%v]", path, err)
		}
	}
}

// CleanDir removes everything in a directory, but not the directory itself
func CleanDir(filePath string) error {
	// If the directory doesn't even exist there's nothing to do, and we do
	// not consider this an error
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	d, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(filePath, name)); err != nil {
			return err
		}
	}
	return nil
}
