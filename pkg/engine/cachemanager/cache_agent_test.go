package cachemanager

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
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/bhojpur/dcp/pkg/engine/util"
)

func TestUpdateCacheAgents(t *testing.T) {
	testcases := map[string]struct {
		desc          string
		initAgents    []string
		cacheAgents   string
		resultAgents  sets.String
		deletedAgents sets.String
	}{
		"two new agents updated": {
			initAgents:    []string{},
			cacheAgents:   "agent1,agent2",
			resultAgents:  sets.NewString(append([]string{"agent1", "agent2"}, util.DefaultCacheAgents...)...),
			deletedAgents: sets.String{},
		},
		"two new agents updated but an old agent deleted": {
			initAgents:    []string{"agent1", "agent2"},
			cacheAgents:   "agent2,agent3",
			resultAgents:  sets.NewString(append([]string{"agent2", "agent3"}, util.DefaultCacheAgents...)...),
			deletedAgents: sets.NewString("agent1"),
		},
		"no agents updated ": {
			initAgents:    []string{"agent1", "agent2"},
			cacheAgents:   "agent1,agent2",
			resultAgents:  sets.NewString(append([]string{"agent1", "agent2"}, util.DefaultCacheAgents...)...),
			deletedAgents: sets.String{},
		},
		"empty agents added ": {
			initAgents:    []string{},
			cacheAgents:   "",
			resultAgents:  sets.NewString(util.DefaultCacheAgents...),
			deletedAgents: sets.String{},
		},
	}
	for k, tt := range testcases {
		t.Run(k, func(t *testing.T) {
			m := &cacheManager{
				cacheAgents: sets.NewString(tt.initAgents...),
			}

			// add agents
			deletedAgents := m.updateCacheAgents(tt.cacheAgents, "")

			if !deletedAgents.Equal(tt.deletedAgents) {
				t.Errorf("Got deleted agents: %v, expect agents: %v", deletedAgents, tt.deletedAgents)
			}

			if !m.cacheAgents.Equal(tt.resultAgents) {
				t.Errorf("Got cache agents: %v, expect agents: %v", m.cacheAgents, tt.resultAgents)
			}
		})
	}
}
