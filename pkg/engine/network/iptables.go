package network

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
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"

	"github.com/bhojpur/dcp/pkg/utils/iptables"
)

type iptablesRule struct {
	pos   iptables.RulePosition
	table iptables.Table
	chain iptables.Chain
	args  []string
}

type IptablesManager struct {
	iptables iptables.Interface
	rules    []iptablesRule
}

func NewIptablesManager(dummyIfIP, dummyIfPort string) *IptablesManager {
	protocol := iptables.ProtocolIpv4
	execer := exec.New()
	iptInterface := iptables.New(execer, protocol)

	im := &IptablesManager{
		iptables: iptInterface,
		rules:    makeupIptablesRules(dummyIfIP, dummyIfPort),
	}

	return im
}

func makeupIptablesRules(ifIP, ifPort string) []iptablesRule {
	return []iptablesRule{
		// skip connection track for traffic from container to 169.254.2.1:10261
		{iptables.Prepend, iptables.Table("raw"), iptables.ChainPrerouting, []string{"-p", "tcp", "--dport", ifPort, "--destination", ifIP, "-j", "NOTRACK"}},
		// skip connection track for traffic from host network to 169.254.2.1:10261
		{iptables.Prepend, iptables.Table("raw"), iptables.ChainOutput, []string{"-p", "tcp", "--dport", ifPort, "--destination", ifIP, "-j", "NOTRACK"}},
		// accept traffic to 169.254.2.1:10261
		{iptables.Prepend, iptables.TableFilter, iptables.ChainInput, []string{"-p", "tcp", "-m", "comment", "--comment", "for container access hub agent", "--dport", ifPort, "--destination", ifIP, "-j", "ACCEPT"}},
		// skip connection track for traffic from 169.254.2.1:10261
		{iptables.Prepend, iptables.Table("raw"), iptables.ChainOutput, []string{"-p", "tcp", "--sport", ifPort, "-s", ifIP, "-j", "NOTRACK"}},
		// accept traffic from 169.254.2.1:10261
		{iptables.Prepend, iptables.TableFilter, iptables.ChainOutput, []string{"-p", "tcp", "--sport", ifPort, "-s", ifIP, "-j", "ACCEPT"}},
		// skip connection track for traffic from container to 127.0.0.1:10261
		{iptables.Prepend, iptables.Table("raw"), iptables.ChainPrerouting, []string{"-p", "tcp", "--dport", ifPort, "--destination", "127.0.0.1", "-j", "NOTRACK"}},
		// skip connection track for traffic from host network to 127.0.0.1:10261
		{iptables.Prepend, iptables.Table("raw"), iptables.ChainOutput, []string{"-p", "tcp", "--dport", ifPort, "--destination", "127.0.0.1", "-j", "NOTRACK"}},
		// accept traffic to 127.0.0.1:10261
		{iptables.Prepend, iptables.TableFilter, iptables.ChainInput, []string{"-p", "tcp", "--dport", ifPort, "--destination", "127.0.0.1", "-j", "ACCEPT"}},
		// skip connection track for traffic from 127.0.0.1:10261
		{iptables.Prepend, iptables.Table("raw"), iptables.ChainOutput, []string{"-p", "tcp", "--sport", ifPort, "-s", "127.0.0.1", "-j", "NOTRACK"}},
		// accept traffic from 127.0.0.1:10261
		{iptables.Prepend, iptables.TableFilter, iptables.ChainOutput, []string{"-p", "tcp", "--sport", ifPort, "-s", "127.0.0.1", "-j", "ACCEPT"}},
	}
}

func (im *IptablesManager) EnsureIptablesRules() error {
	for _, rule := range im.rules {
		_, err := im.iptables.EnsureRule(rule.pos, rule.table, rule.chain, rule.args...)
		if err != nil {
			klog.Errorf("could not ensure iptables rule(%s -t %s %s %s), %v", rule.pos, rule.table, rule.chain, strings.Join(rule.args, ","), err)
			continue
		}
	}
	return nil
}

func (im *IptablesManager) CleanUpIptablesRules() {
	for _, rule := range im.rules {
		err := im.iptables.DeleteRule(rule.table, rule.chain, rule.args...)
		if err != nil {
			klog.Errorf("failed to delete iptables rule(%s -t %s %s %s), %v", rule.pos, rule.table, rule.chain, strings.Join(rule.args, " "), err)
		}
	}
}
