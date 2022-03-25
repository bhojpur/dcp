package healthchecker

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
	"time"
)

type fakeChecker struct {
	healthy  bool
	settings map[string]int
}

// IsHealthy returns healthy status of server
func (fc *fakeChecker) IsHealthy(server *url.URL) bool {
	s := server.String()
	if _, ok := fc.settings[s]; !ok {
		return fc.healthy
	}

	if fc.settings[s] < 0 {
		return fc.healthy
	}

	if fc.settings[s] == 0 {
		return !fc.healthy
	}

	fc.settings[s] = fc.settings[s] - 1
	return fc.healthy
}

func (fc *fakeChecker) Run() {
	return
}

func (fc *fakeChecker) UpdateLastKubeletLeaseReqTime(time.Time) {
	return
}

// NewFakeChecker creates a fake checker
func NewFakeChecker(healthy bool, settings map[string]int) HealthChecker {
	return &fakeChecker{
		settings: settings,
		healthy:  healthy,
	}
}
