package remote

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
	"net/url"
	"testing"

	"github.com/bhojpur/dcp/pkg/engine/healthchecker"
)

type PickBackend struct {
	DeltaRequestsCnt int
	ReturnServer     string
}

func TestRrLoadBalancerAlgo(t *testing.T) {
	testcases := map[string]struct {
		Servers      []string
		PickBackends []PickBackend
	}{
		"no backend servers": {
			Servers: []string{},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: ""},
			},
		},

		"one backend server": {
			Servers: []string{"http://127.0.0.1:8080"},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
			},
		},

		"multi backend server": {
			Servers: []string{"http://127.0.0.1:8080", "http://127.0.0.1:8081", "http://127.0.0.1:8082"},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 2, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 3, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 4, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 4, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 4, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 5, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 5, ReturnServer: "http://127.0.0.1:8080"},
			},
		},
	}

	checker := healthchecker.NewFakeChecker(true, map[string]int{})
	for k, tc := range testcases {
		backends := make([]*RemoteProxy, len(tc.Servers))
		for i := range tc.Servers {
			u, _ := url.Parse(tc.Servers[i])
			backends[i] = &RemoteProxy{
				remoteServer: u,
				checker:      checker,
			}
		}

		rr := &rrLoadBalancerAlgo{
			backends: backends,
		}

		for i := range tc.PickBackends {
			var b *RemoteProxy
			for j := 0; j < tc.PickBackends[i].DeltaRequestsCnt; j++ {
				b = rr.PickOne()
			}

			if len(tc.PickBackends[i].ReturnServer) == 0 {
				if b != nil {
					t.Errorf("%s rr lb pick: expect no backend server, but got %s", k, b.remoteServer.String())
				}
			} else {
				if b == nil {
					t.Errorf("%s rr lb pick: expect backend server: %s, but got no backend server", k, tc.PickBackends[i].ReturnServer)
				} else if b.remoteServer.String() != tc.PickBackends[i].ReturnServer {
					t.Errorf("%s rr lb pick(round %d): expect backend server: %s, but got %s", k, i+1, tc.PickBackends[i].ReturnServer, b.remoteServer.String())
				}
			}
		}
	}
}

func TestRrLoadBalancerAlgoWithReverseHealthy(t *testing.T) {
	testcases := map[string]struct {
		Servers      []string
		PickBackends []PickBackend
	}{
		"multi backend server": {
			Servers: []string{"http://127.0.0.1:8080", "http://127.0.0.1:8081", "http://127.0.0.1:8082"},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
			},
		},
	}

	checker := healthchecker.NewFakeChecker(true, map[string]int{
		"http://127.0.0.1:8080": 1,
		"http://127.0.0.1:8081": 2,
	})
	for k, tc := range testcases {
		backends := make([]*RemoteProxy, len(tc.Servers))
		for i := range tc.Servers {
			u, _ := url.Parse(tc.Servers[i])
			backends[i] = &RemoteProxy{
				remoteServer: u,
				checker:      checker,
			}
		}

		rr := &rrLoadBalancerAlgo{
			backends: backends,
		}

		for i := range tc.PickBackends {
			var b *RemoteProxy
			for j := 0; j < tc.PickBackends[i].DeltaRequestsCnt; j++ {
				b = rr.PickOne()
			}

			if len(tc.PickBackends[i].ReturnServer) == 0 {
				if b != nil {
					t.Errorf("%s rr lb pick: expect no backend server, but got %s", k, b.remoteServer.String())
				}
			} else {
				if b == nil {
					t.Errorf("%s rr lb pick(round %d): expect backend server: %s, but got no backend server", k, i+1, tc.PickBackends[i].ReturnServer)
				} else if b.remoteServer.String() != tc.PickBackends[i].ReturnServer {
					t.Errorf("%s rr lb pick(round %d): expect backend server: %s, but got %s", k, i+1, tc.PickBackends[i].ReturnServer, b.remoteServer.String())
				}
			}
		}
	}
}

func TestPriorityLoadBalancerAlgo(t *testing.T) {
	testcases := map[string]struct {
		Servers      []string
		PickBackends []PickBackend
	}{
		"no backend servers": {
			Servers: []string{},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: ""},
			},
		},

		"one backend server": {
			Servers: []string{"http://127.0.0.1:8080"},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
			},
		},

		"multi backend server": {
			Servers: []string{"http://127.0.0.1:8080", "http://127.0.0.1:8081", "http://127.0.0.1:8082"},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 2, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 3, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 4, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 4, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 4, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 5, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 5, ReturnServer: "http://127.0.0.1:8080"},
			},
		},
	}

	checker := healthchecker.NewFakeChecker(true, map[string]int{})
	for k, tc := range testcases {
		backends := make([]*RemoteProxy, len(tc.Servers))
		for i := range tc.Servers {
			u, _ := url.Parse(tc.Servers[i])
			backends[i] = &RemoteProxy{
				remoteServer: u,
				checker:      checker,
			}
		}

		rr := &priorityLoadBalancerAlgo{
			backends: backends,
		}

		for i := range tc.PickBackends {
			var b *RemoteProxy
			for j := 0; j < tc.PickBackends[i].DeltaRequestsCnt; j++ {
				b = rr.PickOne()
			}

			if len(tc.PickBackends[i].ReturnServer) == 0 {
				if b != nil {
					t.Errorf("%s priority lb pick: expect no backend server, but got %s", k, b.remoteServer.String())
				}
			} else {
				if b == nil {
					t.Errorf("%s priority lb pick: expect backend server: %s, but got no backend server", k, tc.PickBackends[i].ReturnServer)
				} else if b.remoteServer.String() != tc.PickBackends[i].ReturnServer {
					t.Errorf("%s priority lb pick(round %d): expect backend server: %s, but got %s", k, i+1, tc.PickBackends[i].ReturnServer, b.remoteServer.String())
				}
			}
		}
	}
}

func TestPriorityLoadBalancerAlgoWithReverseHealthy(t *testing.T) {
	testcases := map[string]struct {
		Servers      []string
		PickBackends []PickBackend
	}{
		"multi backend server": {
			Servers: []string{"http://127.0.0.1:8080", "http://127.0.0.1:8081", "http://127.0.0.1:8082"},
			PickBackends: []PickBackend{
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8080"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8081"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 1, ReturnServer: "http://127.0.0.1:8082"},
				{DeltaRequestsCnt: 2, ReturnServer: "http://127.0.0.1:8082"},
			},
		},
	}

	checker := healthchecker.NewFakeChecker(true, map[string]int{
		"http://127.0.0.1:8080": 2,
		"http://127.0.0.1:8081": 3})
	for k, tc := range testcases {
		backends := make([]*RemoteProxy, len(tc.Servers))
		for i := range tc.Servers {
			u, _ := url.Parse(tc.Servers[i])
			backends[i] = &RemoteProxy{
				remoteServer: u,
				checker:      checker,
			}
		}

		rr := &priorityLoadBalancerAlgo{
			backends: backends,
		}

		for i := range tc.PickBackends {
			var b *RemoteProxy
			for j := 0; j < tc.PickBackends[i].DeltaRequestsCnt; j++ {
				b = rr.PickOne()
			}

			if len(tc.PickBackends[i].ReturnServer) == 0 {
				if b != nil {
					t.Errorf("%s priority lb pick: expect no backend server, but got %s", k, b.remoteServer.String())
				}
			} else {
				if b == nil {
					t.Errorf("%s priority lb pick: expect backend server: %s, but got no backend server", k, tc.PickBackends[i].ReturnServer)
				} else if b.remoteServer.String() != tc.PickBackends[i].ReturnServer {
					t.Errorf("%s priority lb pick(round %d): expect backend server: %s, but got %s", k, i+1, tc.PickBackends[i].ReturnServer, b.remoteServer.String())
				}
			}
		}
	}
}
