//go:build !windows
// +build !windows

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

import (
	"fmt"
	"os/exec"
	"strings"
)

// OpenRCInitSystem defines openrc
type OpenRCInitSystem struct{}

// ServiceStart tries to start a specific service
func (openrc OpenRCInitSystem) ServiceStart(service string) error {
	args := []string{service, "start"}
	return exec.Command("rc-service", args...).Run()
}

// ServiceStop tries to stop a specific service
func (openrc OpenRCInitSystem) ServiceStop(service string) error {
	args := []string{service, "stop"}
	return exec.Command("rc-service", args...).Run()
}

// ServiceRestart tries to reload the environment and restart the specific service
func (openrc OpenRCInitSystem) ServiceRestart(service string) error {
	args := []string{service, "restart"}
	return exec.Command("rc-service", args...).Run()
}

// ServiceExists ensures the service is defined for this init system.
// openrc writes to stderr if a service is not found or not enabled
// this is in contrast to systemd which only writes to stdout.
// Hence, we use the Combinedoutput, and ignore the error.
func (openrc OpenRCInitSystem) ServiceExists(service string) bool {
	args := []string{service, "status"}
	outBytes, _ := exec.Command("rc-service", args...).CombinedOutput()
	return !strings.Contains(string(outBytes), "does not exist")
}

// ServiceIsEnabled ensures the service is enabled to start on each boot.
func (openrc OpenRCInitSystem) ServiceIsEnabled(service string) bool {
	args := []string{"show", "default"}
	outBytes, _ := exec.Command("rc-update", args...).Output()
	return strings.Contains(string(outBytes), service)
}

// ServiceIsActive ensures the service is running, or attempting to run. (crash looping in the case of kubelet)
func (openrc OpenRCInitSystem) ServiceIsActive(service string) bool {
	args := []string{service, "status"}
	outBytes, _ := exec.Command("rc-service", args...).CombinedOutput()
	outStr := string(outBytes)
	return !strings.Contains(outStr, "stopped") && !strings.Contains(outStr, "does not exist")
}

// EnableCommand return a string describing how to enable a service
func (openrc OpenRCInitSystem) EnableCommand(service string) string {
	return fmt.Sprintf("rc-update add %s default", service)
}

// SystemdInitSystem defines systemd
type SystemdInitSystem struct{}

// EnableCommand return a string describing how to enable a service
func (sysd SystemdInitSystem) EnableCommand(service string) string {
	return fmt.Sprintf("systemctl enable %s.service", service)
}

// reloadSystemd reloads the systemd daemon
func (sysd SystemdInitSystem) reloadSystemd() error {
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}
	return nil
}

// ServiceStart tries to start a specific service
func (sysd SystemdInitSystem) ServiceStart(service string) error {
	// Before we try to start any service, make sure that systemd is ready
	if err := sysd.reloadSystemd(); err != nil {
		return err
	}
	args := []string{"start", service}
	return exec.Command("systemctl", args...).Run()
}

// ServiceRestart tries to reload the environment and restart the specific service
func (sysd SystemdInitSystem) ServiceRestart(service string) error {
	// Before we try to restart any service, make sure that systemd is ready
	if err := sysd.reloadSystemd(); err != nil {
		return err
	}
	args := []string{"restart", service}
	return exec.Command("systemctl", args...).Run()
}

// ServiceStop tries to stop a specific service
func (sysd SystemdInitSystem) ServiceStop(service string) error {
	args := []string{"stop", service}
	return exec.Command("systemctl", args...).Run()
}

// ServiceExists ensures the service is defined for this init system.
func (sysd SystemdInitSystem) ServiceExists(service string) bool {
	args := []string{"status", service}
	outBytes, _ := exec.Command("systemctl", args...).Output()
	output := string(outBytes)
	return !strings.Contains(output, "Loaded: not-found")
}

// ServiceIsEnabled ensures the service is enabled to start on each boot.
func (sysd SystemdInitSystem) ServiceIsEnabled(service string) bool {
	args := []string{"is-enabled", service}
	err := exec.Command("systemctl", args...).Run()
	return err == nil
}

// ServiceIsActive will check is the service is "active". In the case of
// crash looping services (kubelet in our case) status will return as
// "activating", so we will consider this active as well.
func (sysd SystemdInitSystem) ServiceIsActive(service string) bool {
	args := []string{"is-active", service}
	// Ignoring error here, command returns non-0 if in "activating" status:
	outBytes, _ := exec.Command("systemctl", args...).Output()
	output := strings.TrimSpace(string(outBytes))
	if output == "active" || output == "activating" {
		return true
	}
	return false
}

// GetInitSystem returns an InitSystem for the current system, or nil
// if we cannot detect a supported init system.
// This indicates we will skip init system checks, not an error.
func GetInitSystem() (InitSystem, error) {
	// Assume existence of systemctl in path implies this is a systemd system:
	_, err := exec.LookPath("systemctl")
	if err == nil {
		return &SystemdInitSystem{}, nil
	}
	_, err = exec.LookPath("openrc")
	if err == nil {
		return &OpenRCInitSystem{}, nil
	}

	return nil, fmt.Errorf("no supported init system detected, skipping checking for services")
}
