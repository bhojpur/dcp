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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"

	"github.com/bhojpur/dcp/pkg/cloud/flock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// Compile-time variable
var existingServer = "False"

const lockFile = "/tmp/dcp-test.lock"

type DcpServer struct {
	cmd     *exec.Cmd
	scanner *bufio.Scanner
}

func findDcpExecutable() string {
	// if running on an existing cluster, it maybe installed via dcp.service
	// or run manually from dist/artifacts/dcp
	if IsExistingServer() {
		dcpBin, err := exec.LookPath("dcp")
		if err == nil {
			return dcpBin
		}
	}
	dcpBin := "dist/artifacts/dcp"
	i := 0
	for ; i < 20; i++ {
		_, err := os.Stat(dcpBin)
		if err != nil {
			dcpBin = "../" + dcpBin
			continue
		}
		break
	}
	if i == 20 {
		logrus.Fatal("Unable to find Bhojpur DCP executable")
	}
	return dcpBin
}

// IsRoot return true if the user is root (UID 0)
func IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}

func IsExistingServer() bool {
	return existingServer == "True"
}

// DcpCmd launches the provided Bhojpur DCP command via exec. Command blocks until finished.
// Command output from both Stderr and Stdout is provided via string. Input can
// be a single string with space separated args, or multiple string args
//   cmdEx1, err := DcpCmd("etcd-snapshot", "ls")
//   cmdEx2, err := DcpCmd("kubectl get pods -A")
//   cmdEx2, err := DcpCmd("kubectl", "get", "pods", "-A")
func DcpCmd(inputArgs ...string) (string, error) {
	if !IsRoot() {
		return "", errors.New("integration tests must be run as sudo/root")
	}
	dcpBin := findDcpExecutable()
	var dcpCmd []string
	for _, arg := range inputArgs {
		dcpCmd = append(dcpCmd, strings.Fields(arg)...)
	}
	cmd := exec.Command(dcpBin, dcpCmd...)
	byteOut, err := cmd.CombinedOutput()
	return string(byteOut), err
}

func contains(source []string, target string) bool {
	for _, s := range source {
		if s == target {
			return true
		}
	}
	return false
}

// ServerArgsPresent checks if the given arguments are found in the running Bhojpur DCP server
func ServerArgsPresent(neededArgs []string) bool {
	currentArgs := DcpServerArgs()
	for _, arg := range neededArgs {
		if !contains(currentArgs, arg) {
			return false
		}
	}
	return true
}

// DcpServerArgs returns the list of arguments that the Bhojpur DCP server launched with
func DcpServerArgs() []string {
	results, err := DcpCmd("kubectl", "get", "nodes", "-o", `jsonpath='{.items[0].metadata.annotations.bhojpur\.net/node-args}'`)
	if err != nil {
		return nil
	}
	res := strings.ReplaceAll(results, "'", "")
	var args []string
	if err := json.Unmarshal([]byte(res), &args); err != nil {
		logrus.Error(err)
		return nil
	}
	return args
}

func FindStringInCmdAsync(scanner *bufio.Scanner, target string) bool {
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), target) {
			return true
		}
	}
	return false
}

func DcpTestLock() (int, error) {
	logrus.Info("waiting to get test lock")
	return flock.Acquire(lockFile)
}

// DcpStartServer acquires an exclusive lock on a temporary file, then launches a Bhojpur DCP cluster
// with the provided arguments. Subsequent/parallel calls to this function will block until
// the original lock is cleared using DcpKillServer
func DcpStartServer(inputArgs ...string) (*DcpServer, error) {
	if !IsRoot() {
		return nil, errors.New("integration tests must be run as sudo/root")
	}

	var cmdArgs []string
	for _, arg := range inputArgs {
		cmdArgs = append(cmdArgs, strings.Fields(arg)...)
	}
	dcpBin := findDcpExecutable()
	dcpCmd := append([]string{"server"}, cmdArgs...)
	cmd := exec.Command(dcpBin, dcpCmd...)
	// Give the server a new group id so we can kill it and its children later
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmdOut, _ := cmd.StderrPipe()
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	return &DcpServer{cmd, bufio.NewScanner(cmdOut)}, err
}

// DcpKillServer terminates the running Bhojpur DCP server and its children
// and unlocks the file for other tests
func DcpKillServer(server *DcpServer) error {
	pgid, err := syscall.Getpgid(server.cmd.Process.Pid)
	if err != nil {
		if errors.Is(err, syscall.ESRCH) {
			logrus.Warnf("Unable to kill Bhojpur DCP server: %v", err)
			return nil
		}
		return errors.Wrap(err, "failed to find Bhojpur DCP process group")
	}
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		return errors.Wrap(err, "failed to kill Bhojpur DCP process group")
	}
	if err := server.cmd.Process.Kill(); err != nil {
		return errors.Wrap(err, "failed to kill Bhojpur DCP process")
	}
	if _, err = server.cmd.Process.Wait(); err != nil {
		return errors.Wrap(err, "failed to wait for Bhojpur DCP process exit")
	}
	return nil
}

// DcpCleanup attempts to cleanup networking and files leftover from an integration test
// this is similar to the dcp-killall.sh script, but we dynamically generate that on
// install, so we don't have access to it in testing.
func DcpCleanup(dcpTestLock int, dataDir string) error {
	if cni0Link, err := netlink.LinkByName("cni0"); err == nil {
		links, _ := netlink.LinkList()
		for _, link := range links {
			if link.Attrs().MasterIndex == cni0Link.Attrs().Index {
				netlink.LinkDel(link)
			}
		}
		netlink.LinkDel(cni0Link)
	}

	if flannel1, err := netlink.LinkByName("flannel.1"); err == nil {
		netlink.LinkDel(flannel1)
	}
	if flannelV6, err := netlink.LinkByName("flannel-v6.1"); err == nil {
		netlink.LinkDel(flannelV6)
	}
	if dataDir == "" {
		dataDir = "/var/lib/bhojpur/dcp"
	}
	if err := os.RemoveAll(dataDir); err != nil {
		return err
	}
	return flock.Release(dcpTestLock)
}

// RunCommand Runs command on the host
func RunCommand(cmd string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	var out bytes.Buffer
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	return out.String(), nil
}
