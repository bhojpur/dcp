package initsystem

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

// InitSystem is the interface that describe behaviors of an init system
type InitSystem interface {
	// return a string describing how to enable a service
	EnableCommand(service string) string

	// ServiceStart tries to start a specific service
	ServiceStart(service string) error

	// ServiceStop tries to stop a specific service
	ServiceStop(service string) error

	// ServiceRestart tries to reload the environment and restart the specific service
	ServiceRestart(service string) error

	// ServiceExists ensures the service is defined for this init system.
	ServiceExists(service string) bool

	// ServiceIsEnabled ensures the service is enabled to start on each boot.
	ServiceIsEnabled(service string) bool

	// ServiceIsActive ensures the service is running, or attempting to run. (crash looping in the case of kubelet)
	ServiceIsActive(service string) bool
}
