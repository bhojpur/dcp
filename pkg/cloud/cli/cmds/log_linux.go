//go:build linux && cgo
// +build linux,cgo

package cmds

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
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bhojpur/dcp/pkg/cloud/version"
	systemd "github.com/coreos/go-systemd/daemon"
	"github.com/erikdubbelboer/gspt"
	"github.com/natefinch/lumberjack"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// forkIfLoggingOrReaping handles forking off the actual Bhojpur DCP process if it is necessary to
// capture log output, or reap child processes. Reaping is only necessary when running
// as pid 1.
func forkIfLoggingOrReaping() error {
	var stdout, stderr io.Writer = os.Stdout, os.Stderr
	enableLogRedirect := LogConfig.LogFile != "" && os.Getenv("_DCP_LOG_REEXEC_") == ""
	enableReaping := os.Getpid() == 1

	if enableLogRedirect {
		var l io.Writer = &lumberjack.Logger{
			Filename:   LogConfig.LogFile,
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
		if LogConfig.AlsoLogToStderr {
			l = io.MultiWriter(l, os.Stderr)
		}
		stdout = l
		stderr = l
	}

	if enableLogRedirect || enableReaping {
		gspt.SetProcTitle(os.Args[0] + " init")

		pwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "failed to get working directory")
		}

		if enableReaping {
			// If we're running as pid 1 we need to reap child processes or defunct containerd-shim
			// child processes will accumulate.
			unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(1), 0, 0, 0)
			go reapChildren()
		}

		args := append([]string{version.Program}, os.Args[1:]...)
		env := append(os.Environ(), "_DCP_LOG_REEXEC_=true", "NOTIFY_SOCKET=")
		cmd := &exec.Cmd{
			Path:   "/proc/self/exe",
			Dir:    pwd,
			Args:   args,
			Env:    env,
			Stdin:  os.Stdin,
			Stdout: stdout,
			Stderr: stderr,
			SysProcAttr: &syscall.SysProcAttr{
				Pdeathsig: unix.SIGTERM,
			},
		}
		if err := cmd.Start(); err != nil {
			return err
		}

		// The child process won't be allowed to notify, so we send one for it as soon as it's started,
		// and then wait for it to exit and pass along the exit code.
		systemd.SdNotify(true, "READY=1\n")
		cmd.Wait()
		os.Exit(cmd.ProcessState.ExitCode())
	}
	return nil
}

//reapChildren calls Wait4 whenever SIGCHLD is received
func reapChildren() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGCHLD)
	for {
		select {
		case <-sigs:
		}
		for {
			var wstatus syscall.WaitStatus
			_, err := syscall.Wait4(-1, &wstatus, 0, nil)
			for err == syscall.EINTR {
				_, err = syscall.Wait4(-1, &wstatus, 0, nil)
			}
			if err == nil || err == syscall.ECHILD {
				break
			}
		}
	}
}
