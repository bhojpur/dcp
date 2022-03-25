package options

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
	"github.com/spf13/pflag"
)

// AppOptions is the main settings for the app-manager
type AppOptions struct {
	MetricsAddr             string
	PprofAddr               string
	HealthProbeAddr         string
	EnableLeaderElection    bool
	EnablePprof             bool
	LeaderElectionNamespace string
	Namespace               string
	CreateDefaultPool       bool
	Version                 bool
}

// NewAppOptions creates a new AppOptions with a default config.
func NewAppOptions() *AppOptions {
	o := &AppOptions{
		MetricsAddr:             ":8080",
		PprofAddr:               ":8090",
		HealthProbeAddr:         ":8000",
		EnableLeaderElection:    true,
		EnablePprof:             false,
		LeaderElectionNamespace: "kube-system",
		Namespace:               "",
		CreateDefaultPool:       false,
	}

	return o
}

// ValidateOptions validates AppOptions
func ValidateOptions(options *AppOptions) error {
	// TODO
	return nil
}

// AddFlags returns flags for a specific Engine by section name
func (o *AppOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.MetricsAddr, "metrics-addr", o.MetricsAddr, "The address the metric endpoint binds to.")
	fs.StringVar(&o.PprofAddr, "pprof-addr", o.PprofAddr, "The address the pprof binds to.")
	fs.StringVar(&o.HealthProbeAddr, "health-probe-addr", o.HealthProbeAddr, "The address the healthz/readyz endpoint binds to.")
	fs.BoolVar(&o.EnableLeaderElection, "enable-leader-election", o.EnableLeaderElection, "Whether you need to enable leader election.")
	fs.BoolVar(&o.EnablePprof, "enable-pprof", o.EnablePprof, "Enable pprof for controller manager.")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-election-namespace", o.LeaderElectionNamespace, "This determines the namespace in which the leader election configmap will be created, it will use in-cluster namespace if empty.")
	fs.StringVar(&o.Namespace, "namespace", o.Namespace, "Namespace if specified restricts the manager's cache to watch objects in the desired namespace. Defaults to all namespaces.")
	fs.BoolVar(&o.CreateDefaultPool, "create-default-pool", o.CreateDefaultPool, "Create default cloud/edge pools if indicated.")
	fs.BoolVar(&o.Version, "version", o.Version, "print the version information.")
}
