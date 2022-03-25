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
	"errors"
	"math/rand"
	"reflect"
)

func (lb *LoadBalancer) setServers(serverAddresses []string) bool {
	serverAddresses, hasOriginalServer := sortServers(serverAddresses, lb.defaultServerAddress)
	if len(serverAddresses) == 0 {
		return false
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if reflect.DeepEqual(serverAddresses, lb.ServerAddresses) {
		return false
	}

	lb.ServerAddresses = serverAddresses
	lb.randomServers = append([]string{}, lb.ServerAddresses...)
	rand.Shuffle(len(lb.randomServers), func(i, j int) {
		lb.randomServers[i], lb.randomServers[j] = lb.randomServers[j], lb.randomServers[i]
	})
	if !hasOriginalServer {
		lb.randomServers = append(lb.randomServers, lb.defaultServerAddress)
	}
	lb.currentServerAddress = lb.randomServers[0]
	lb.nextServerIndex = 1

	return true
}

func (lb *LoadBalancer) nextServer(failedServer string) (string, error) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if len(lb.randomServers) == 0 {
		return "", errors.New("No servers in load balancer proxy list")
	}
	if len(lb.randomServers) == 1 {
		return lb.currentServerAddress, nil
	}
	if failedServer != lb.currentServerAddress {
		return lb.currentServerAddress, nil
	}
	if lb.nextServerIndex >= len(lb.randomServers) {
		lb.nextServerIndex = 0
	}

	lb.currentServerAddress = lb.randomServers[lb.nextServerIndex]
	lb.nextServerIndex++

	return lb.currentServerAddress, nil
}
