//go:build linux
// +build linux

package iptables

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
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/exec"
)

// We can't use the normal FakeExec because we don't know precisely how many times the
// Monitor thread will do its checks, and we don't know precisely how its iptables calls
// will interleave with the main thread's. So we use our own fake Exec implementation that
// implements a minimal iptables interface. This will need updates as iptables.runner
// changes its use of Exec.
type monitorFakeExec struct {
	sync.Mutex

	tables map[string]sets.String

	block      bool
	wasBlocked bool
}

func newMonitorFakeExec() *monitorFakeExec {
	tables := make(map[string]sets.String)
	tables["mangle"] = sets.NewString()
	tables["filter"] = sets.NewString()
	tables["nat"] = sets.NewString()
	return &monitorFakeExec{tables: tables}
}

func (mfe *monitorFakeExec) blockIPTables(block bool) {
	mfe.Lock()
	defer mfe.Unlock()

	mfe.block = block
}

func (mfe *monitorFakeExec) getWasBlocked() bool {
	mfe.Lock()
	defer mfe.Unlock()

	wasBlocked := mfe.wasBlocked
	mfe.wasBlocked = false
	return wasBlocked
}

func (mfe *monitorFakeExec) Command(cmd string, args ...string) exec.Cmd {
	return &monitorFakeCmd{mfe: mfe, cmd: cmd, args: args}
}

func (mfe *monitorFakeExec) CommandContext(ctx context.Context, cmd string, args ...string) exec.Cmd {
	return mfe.Command(cmd, args...)
}

func (mfe *monitorFakeExec) LookPath(file string) (string, error) {
	return file, nil
}

type monitorFakeCmd struct {
	mfe  *monitorFakeExec
	cmd  string
	args []string
}

func (mfc *monitorFakeCmd) CombinedOutput() ([]byte, error) {
	if mfc.cmd == cmdIPTablesRestore {
		// Only used for "iptables-restore --version", and the result doesn't matter
		return []byte{}, nil
	} else if mfc.cmd != cmdIPTables {
		panic("bad command " + mfc.cmd)
	}

	if len(mfc.args) == 1 && mfc.args[0] == "--version" {
		return []byte("iptables v1.6.2"), nil
	}

	if len(mfc.args) != 8 || mfc.args[0] != WaitString || mfc.args[1] != WaitSecondsValue || mfc.args[2] != WaitIntervalString || mfc.args[3] != WaitIntervalUsecondsValue || mfc.args[6] != "-t" {
		panic(fmt.Sprintf("bad args %#v", mfc.args))
	}
	op := operation(mfc.args[4])
	chainName := mfc.args[5]
	tableName := mfc.args[7]

	mfc.mfe.Lock()
	defer mfc.mfe.Unlock()

	table := mfc.mfe.tables[tableName]
	if table == nil {
		return []byte{}, fmt.Errorf("no such table %q", tableName)
	}

	// For ease-of-testing reasons, blockIPTables blocks create and list, but not delete
	if mfc.mfe.block && op != opDeleteChain {
		mfc.mfe.wasBlocked = true
		return []byte{}, exec.CodeExitError{Code: 4, Err: fmt.Errorf("could not get xtables.lock, etc")}
	}

	switch op {
	case opCreateChain:
		if !table.Has(chainName) {
			table.Insert(chainName)
		}
		return []byte{}, nil
	case opListChain:
		if table.Has(chainName) {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("no such chain %q", chainName)
	case opDeleteChain:
		table.Delete(chainName)
		return []byte{}, nil
	default:
		panic("should not be reached")
	}
}

func (mfc *monitorFakeCmd) SetStdin(in io.Reader) {
	// Used by getIPTablesRestoreVersionString(), can be ignored
}

func (mfc *monitorFakeCmd) Run() error {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) Output() ([]byte, error) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) SetDir(dir string) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) SetStdout(out io.Writer) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) SetStderr(out io.Writer) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) SetEnv(env []string) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) StdoutPipe() (io.ReadCloser, error) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) StderrPipe() (io.ReadCloser, error) {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) Start() error {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) Wait() error {
	panic("should not be reached")
}
func (mfc *monitorFakeCmd) Stop() {
	panic("should not be reached")
}

func TestIPTablesMonitor(t *testing.T) {
	mfe := newMonitorFakeExec()
	ipt := New(mfe, ProtocolIpv4)

	var reloads uint32
	stopCh := make(chan struct{})

	canary := Chain("MONITOR-TEST-CANARY")
	tables := []Table{TableMangle, TableFilter, TableNAT}
	go ipt.Monitor(canary, tables, func() {
		if !ensureNoChains(mfe) {
			t.Errorf("reload called while canaries still exist")
		}
		atomic.AddUint32(&reloads, 1)
	}, 100*time.Millisecond, stopCh)

	// Monitor should create canary chains quickly
	if err := waitForChains(mfe, canary, tables); err != nil {
		t.Errorf("failed to create iptables canaries: %v", err)
	}

	if err := waitForReloads(&reloads, 0); err != nil {
		t.Errorf("got unexpected reloads: %v", err)
	}

	// If we delete all of the chains, it should reload
	ipt.DeleteChain(TableMangle, canary)
	ipt.DeleteChain(TableFilter, canary)
	ipt.DeleteChain(TableNAT, canary)

	if err := waitForReloads(&reloads, 1); err != nil {
		t.Errorf("got unexpected number of reloads after flush: %v", err)
	}
	if err := waitForChains(mfe, canary, tables); err != nil {
		t.Errorf("failed to create iptables canaries: %v", err)
	}

	// If we delete two chains, it should not reload yet
	ipt.DeleteChain(TableMangle, canary)
	ipt.DeleteChain(TableFilter, canary)

	if err := waitForNoReload(&reloads, 1); err != nil {
		t.Errorf("got unexpected number of reloads after partial flush: %v", err)
	}

	// Now ensure that "iptables -L" will get an error about the xtables.lock, and
	// delete the last chain. The monitor should not reload, because it can't actually
	// tell if the chain was deleted or not.
	mfe.blockIPTables(true)
	ipt.DeleteChain(TableNAT, canary)
	if err := waitForBlocked(mfe); err != nil {
		t.Errorf("failed waiting for monitor to be blocked from monitoring: %v", err)
	}

	// After unblocking the monitor, it should now reload
	mfe.blockIPTables(false)

	if err := waitForReloads(&reloads, 2); err != nil {
		t.Errorf("got unexpected number of reloads after slow flush: %v", err)
	}
	if err := waitForChains(mfe, canary, tables); err != nil {
		t.Errorf("failed to create iptables canaries: %v", err)
	}

	// If we close the stop channel, it should stop running
	close(stopCh)

	if err := waitForNoReload(&reloads, 2); err != nil {
		t.Errorf("got unexpected number of reloads after stop: %v", err)
	}
	if !ensureNoChains(mfe) {
		t.Errorf("canaries still exist after stopping monitor")
	}

	// If we create a new monitor while the iptables lock is held, it will
	// retry creating canaries until it succeeds

	stopCh = make(chan struct{})
	_ = mfe.getWasBlocked()
	mfe.blockIPTables(true)
	go ipt.Monitor(canary, tables, func() {
		if !ensureNoChains(mfe) {
			t.Errorf("reload called while canaries still exist")
		}
		atomic.AddUint32(&reloads, 1)
	}, 100*time.Millisecond, stopCh)

	// Monitor should not have created canaries yet
	if !ensureNoChains(mfe) {
		t.Errorf("canary created while iptables blocked")
	}

	if err := waitForBlocked(mfe); err != nil {
		t.Errorf("failed waiting for monitor to fail creating canaries: %v", err)
	}

	mfe.blockIPTables(false)
	if err := waitForChains(mfe, canary, tables); err != nil {
		t.Errorf("failed to create iptables canaries: %v", err)
	}

	close(stopCh)
}

func waitForChains(mfe *monitorFakeExec, canary Chain, tables []Table) error {
	return utilwait.PollImmediate(100*time.Millisecond, time.Second, func() (bool, error) {
		mfe.Lock()
		defer mfe.Unlock()

		for _, table := range tables {
			if !mfe.tables[string(table)].Has(string(canary)) {
				return false, nil
			}
		}
		return true, nil
	})
}

func ensureNoChains(mfe *monitorFakeExec) bool {
	mfe.Lock()
	defer mfe.Unlock()
	return mfe.tables["mangle"].Len() == 0 &&
		mfe.tables["filter"].Len() == 0 &&
		mfe.tables["nat"].Len() == 0
}

func waitForReloads(reloads *uint32, expected uint32) error {
	if atomic.LoadUint32(reloads) < expected {
		utilwait.PollImmediate(100*time.Millisecond, time.Second, func() (bool, error) {
			return atomic.LoadUint32(reloads) >= expected, nil
		})
	}
	got := atomic.LoadUint32(reloads)
	if got != expected {
		return fmt.Errorf("expected %d, got %d", expected, got)
	}
	return nil
}

func waitForNoReload(reloads *uint32, expected uint32) error {
	utilwait.PollImmediate(50*time.Millisecond, 250*time.Millisecond, func() (bool, error) {
		return atomic.LoadUint32(reloads) > expected, nil
	})

	got := atomic.LoadUint32(reloads)
	if got != expected {
		return fmt.Errorf("expected %d, got %d", expected, got)
	}
	return nil
}

func waitForBlocked(mfe *monitorFakeExec) error {
	return utilwait.PollImmediate(100*time.Millisecond, time.Second, func() (bool, error) {
		blocked := mfe.getWasBlocked()
		return blocked, nil
	})
}
