package loadbalancer

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
	"encoding/json"
	"io/ioutil"

	"github.com/bhojpur/dcp/pkg/cloud/agent/util"
)

func (lb *LoadBalancer) writeConfig() error {
	configOut, err := json.MarshalIndent(lb, "", "  ")
	if err != nil {
		return err
	}
	return util.WriteFile(lb.configFile, string(configOut))
}

func (lb *LoadBalancer) updateConfig() error {
	writeConfig := true
	if configBytes, err := ioutil.ReadFile(lb.configFile); err == nil {
		config := &LoadBalancer{}
		if err := json.Unmarshal(configBytes, config); err == nil {
			if config.ServerURL == lb.ServerURL {
				writeConfig = false
				lb.setServers(config.ServerAddresses)
			}
		}
	}
	if writeConfig {
		if err := lb.writeConfig(); err != nil {
			return err
		}
	}
	return nil
}
