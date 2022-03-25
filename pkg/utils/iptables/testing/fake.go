package testing

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
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/bhojpur/dcp/pkg/utils/iptables"
)

const (
	// Destination represents the destination address flag
	Destination = "-d "
	// Source represents the source address flag
	Source = "-s "
	// DPort represents the destination port flag
	DPort = "--dport "
	// Protocol represents the protocol flag
	Protocol = "-p "
	// Jump represents jump flag specifies the jump target
	Jump = "-j "
	// Reject specifies the reject target
	Reject = "REJECT"
	// ToDest represents the flag used to specify the destination address in DNAT
	ToDest = "--to-destination "
	// Recent represents the sub-command recent that allows to dynamically create list of IP address to match against
	Recent = "recent "
	// MatchSet represents the flag which match packets against the specified set
	MatchSet = "--match-set "
	// SrcType represents the --src-type flag which matches if the source address is of given type
	SrcType = "--src-type "
	// Masquerade represents the target that is used in nat table.
	Masquerade = "MASQUERADE "
)

// Rule holds a map of rules.
type Rule map[string]string

// FakeIPTables is no-op implementation of iptables Interface.
type FakeIPTables struct {
	hasRandomFully bool
	Lines          []byte
	ipv6           bool
}

// NewFake returns a no-op iptables.Interface
func NewFake() *FakeIPTables {
	return &FakeIPTables{}
}

// NewIpv6Fake returns a no-op iptables.Interface with IsIPv6() == true
func NewIpv6Fake() *FakeIPTables {
	return &FakeIPTables{ipv6: true}
}

// SetHasRandomFully is part of iptables.Interface
func (f *FakeIPTables) SetHasRandomFully(can bool) *FakeIPTables {
	f.hasRandomFully = can
	return f
}

// EnsureChain is part of iptables.Interface
func (*FakeIPTables) EnsureChain(table iptables.Table, chain iptables.Chain) (bool, error) {
	return true, nil
}

// FlushChain is part of iptables.Interface
func (*FakeIPTables) FlushChain(table iptables.Table, chain iptables.Chain) error {
	return nil
}

// DeleteChain is part of iptables.Interface
func (*FakeIPTables) DeleteChain(table iptables.Table, chain iptables.Chain) error {
	return nil
}

// EnsureRule is part of iptables.Interface
func (*FakeIPTables) EnsureRule(position iptables.RulePosition, table iptables.Table, chain iptables.Chain, args ...string) (bool, error) {
	return true, nil
}

// DeleteRule is part of iptables.Interface
func (*FakeIPTables) DeleteRule(table iptables.Table, chain iptables.Chain, args ...string) error {
	return nil
}

// IsIpv6 is part of iptables.Interface
func (f *FakeIPTables) IsIpv6() bool {
	return f.ipv6
}

// Save is part of iptables.Interface
func (f *FakeIPTables) Save(table iptables.Table) ([]byte, error) {
	lines := make([]byte, len(f.Lines))
	copy(lines, f.Lines)
	return lines, nil
}

// SaveInto is part of iptables.Interface
func (f *FakeIPTables) SaveInto(table iptables.Table, buffer *bytes.Buffer) error {
	buffer.Write(f.Lines)
	return nil
}

// Restore is part of iptables.Interface
func (*FakeIPTables) Restore(table iptables.Table, data []byte, flush iptables.FlushFlag, counters iptables.RestoreCountersFlag) error {
	return nil
}

// RestoreAll is part of iptables.Interface
func (f *FakeIPTables) RestoreAll(data []byte, flush iptables.FlushFlag, counters iptables.RestoreCountersFlag) error {
	f.Lines = data
	return nil
}

// Monitor is part of iptables.Interface
func (f *FakeIPTables) Monitor(canary iptables.Chain, tables []iptables.Table, reloadFunc func(), interval time.Duration, stopCh <-chan struct{}) {
}

func getToken(line, separator string) string {
	tokens := strings.Split(line, separator)
	if len(tokens) == 2 {
		return strings.Split(tokens[1], " ")[0]
	}
	return ""
}

// GetRules is part of iptables.Interface
func (f *FakeIPTables) GetRules(chainName string) (rules []Rule) {
	for _, l := range strings.Split(string(f.Lines), "\n") {
		if strings.Contains(l, fmt.Sprintf("-A %v", chainName)) {
			newRule := Rule(map[string]string{})
			for _, arg := range []string{Destination, Source, DPort, Protocol, Jump, ToDest, Recent, MatchSet, SrcType, Masquerade} {
				tok := getToken(l, arg)
				if tok != "" {
					newRule[arg] = tok
				}
			}
			rules = append(rules, newRule)
		}
	}
	return
}

// HasRandomFully is part of iptables.Interface
func (f *FakeIPTables) HasRandomFully() bool {
	return f.hasRandomFully
}

var _ = iptables.Interface(&FakeIPTables{})
