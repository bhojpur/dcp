package components

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
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/pkg/errors"
	utilsexec "k8s.io/utils/exec"
)

const (
	dockerSocket     = "/var/run/docker.sock" // The Docker socket is not CRI compatible
	containerdSocket = "/run/containerd/containerd.sock"
	// DefaultDockerCRISocket defines the default Docker CRI socket
	DefaultDockerCRISocket = "/var/run/dockershim.sock"

	// PullImageRetry specifies how many times ContainerRuntime retries when pulling image failed
	PullImageRetry = 5
)

// ContainerRuntime is an interface for working with container runtimes
type ContainerRuntimeForImage interface {
	IsDocker() bool
	PullImage(image string) error
	ImageExists(image string) (bool, error)
}

// CRIRuntime is a struct that interfaces with the CRI
type CRIRuntime struct {
	exec      utilsexec.Interface
	criSocket string
}

// DockerRuntime is a struct that interfaces with the Docker daemon
type DockerRuntime struct {
	exec utilsexec.Interface
}

// NewContainerRuntime sets up and returns a ContainerRuntime struct
func NewContainerRuntimeForImage(execer utilsexec.Interface, criSocket string) (ContainerRuntimeForImage, error) {
	var toolName string
	var runtime ContainerRuntimeForImage

	if criSocket != DefaultDockerCRISocket {
		toolName = "crictl"
		// !!! temporary work around crictl warning:
		// Using "/var/run/crio/crio.sock" as endpoint is deprecated,
		// please consider using full url format "unix:///var/run/crio/crio.sock"
		if filepath.IsAbs(criSocket) && goruntime.GOOS != "windows" {
			criSocket = "unix://" + criSocket
		}
		runtime = &CRIRuntime{execer, criSocket}
	} else {
		toolName = "docker"
		runtime = &DockerRuntime{execer}
	}

	if _, err := execer.LookPath(toolName); err != nil {
		return nil, errors.Wrapf(err, "%s is required for container runtime", toolName)
	}
	return runtime, nil
}

// IsDocker returns true if the runtime is docker
func (runtime *CRIRuntime) IsDocker() bool {
	return false
}

// IsDocker returns true if the runtime is docker
func (runtime *DockerRuntime) IsDocker() bool {
	return true
}

// PullImage pulls the image
func (runtime *CRIRuntime) PullImage(image string) error {
	var err error
	var out []byte
	for i := 0; i < PullImageRetry; i++ {
		out, err = runtime.exec.Command("crictl", "-r", runtime.criSocket, "pull", image).CombinedOutput()
		if err == nil {
			return nil
		}
	}
	return errors.Wrapf(err, "output: %s, error", out)
}

// PullImage pulls the image
func (runtime *DockerRuntime) PullImage(image string) error {
	var err error
	var out []byte
	for i := 0; i < PullImageRetry; i++ {
		out, err = runtime.exec.Command("docker", "pull", image).CombinedOutput()
		if err == nil {
			return nil
		}
	}
	return errors.Wrapf(err, "output: %s, error", out)
}

// ImageExists checks to see if the image exists on the system
func (runtime *CRIRuntime) ImageExists(image string) (bool, error) {
	err := runtime.exec.Command("crictl", "-r", runtime.criSocket, "inspecti", image).Run()
	return err == nil, nil
}

// ImageExists checks to see if the image exists on the system
func (runtime *DockerRuntime) ImageExists(image string) (bool, error) {
	err := runtime.exec.Command("docker", "inspect", image).Run()
	return err == nil, nil
}

// detectCRISocketImpl is separated out only for test purposes, DON'T call it directly, use DetectCRISocket instead
func detectCRISocketImpl(isSocket func(string) bool) (string, error) {
	foundCRISockets := []string{}
	knownCRISockets := []string{
		// Docker and containerd sockets are special cased below, hence not to be included here
		"/var/run/crio/crio.sock",
	}

	if isSocket(dockerSocket) {
		// the path in dockerSocket is not CRI compatible, hence we should replace it with a CRI compatible socket
		foundCRISockets = append(foundCRISockets, DefaultDockerCRISocket)
	} else if isSocket(containerdSocket) {
		// Docker 18.09 gets bundled together with containerd, thus having both dockerSocket and containerdSocket present.
		// For compatibility reasons, we use the containerd socket only if Docker is not detected.
		foundCRISockets = append(foundCRISockets, containerdSocket)
	}

	for _, socket := range knownCRISockets {
		if isSocket(socket) {
			foundCRISockets = append(foundCRISockets, socket)
		}
	}

	switch len(foundCRISockets) {
	case 0:
		// Fall back to Docker if no CRI is detected, we can error out later on if we need it
		return DefaultDockerCRISocket, nil
	case 1:
		// Precisely one CRI found, use that
		return foundCRISockets[0], nil
	default:
		// Multiple CRIs installed?
		return "", errors.Errorf("Found multiple CRI sockets, please use --cri-socket to select one: %s", strings.Join(foundCRISockets, ", "))
	}

}

// isExistingSocket checks if path exists and is domain socket
func isExistingSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.Mode()&os.ModeSocket != 0
}

// DetectCRISocket uses a list of known CRI sockets to detect one. If more than one or none is discovered, an error is returned.
func DetectCRISocket() (string, error) {
	return detectCRISocketImpl(isExistingSocket)
}
