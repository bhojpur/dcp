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
	nodelifecycleconfig "k8s.io/kube-controller-manager/config/v1alpha1"
)

// NodeLifecycleControllerOptions holds the NodeLifecycleController options.
type NodeLifecycleControllerOptions struct {
	*nodelifecycleconfig.NodeLifecycleControllerConfiguration
}

// AddFlags adds flags related to NodeLifecycleController for controller manager to the specified FlagSet.
func (o *NodeLifecycleControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.DurationVar(&o.NodeStartupGracePeriod.Duration, "node-startup-grace-period", o.NodeStartupGracePeriod.Duration,
		"Amount of time which we allow starting Node to be unresponsive before marking it unhealthy.")
	fs.DurationVar(&o.NodeMonitorGracePeriod.Duration, "node-monitor-grace-period", o.NodeMonitorGracePeriod.Duration,
		"Amount of time which we allow running Node to be unresponsive before marking it unhealthy. "+
			"Must be N times more than kubelet's nodeStatusUpdateFrequency, "+
			"where N means number of retries allowed for kubelet to post node status.")
	fs.DurationVar(&o.PodEvictionTimeout.Duration, "pod-eviction-timeout", o.PodEvictionTimeout.Duration, "The grace period for deleting pods on failed nodes.")
	fs.Float32Var(&o.NodeEvictionRate, "node-eviction-rate", 0.1, "Number of nodes per second on which pods are deleted in case of node failure when a zone is healthy (see --unhealthy-zone-threshold for definition of healthy/unhealthy). Zone refers to entire cluster in non-multizone clusters.")
	fs.Float32Var(&o.SecondaryNodeEvictionRate, "secondary-node-eviction-rate", 0.01, "Number of nodes per second on which pods are deleted in case of node failure when a zone is unhealthy (see --unhealthy-zone-threshold for definition of healthy/unhealthy). Zone refers to entire cluster in non-multizone clusters. This value is implicitly overridden to 0 if the cluster size is smaller than --large-cluster-size-threshold.")
	fs.Int32Var(&o.LargeClusterSizeThreshold, "large-cluster-size-threshold", 50, "Number of nodes from which NodeController treats the cluster as large for the eviction logic purposes. --secondary-node-eviction-rate is implicitly overridden to 0 for clusters this size or smaller.")
	fs.Float32Var(&o.UnhealthyZoneThreshold, "unhealthy-zone-threshold", 0.55, "Fraction of Nodes in a zone which needs to be not Ready (minimum 3) for zone to be treated as unhealthy. ")
	fs.BoolVar(o.EnableTaintManager, "enable-taint-manager", *o.EnableTaintManager, "WARNING: Beta feature. If set to true enables NoExecute Taints and will evict all not-tolerating Pod running on Nodes tainted with this kind of Taints.")
}

// ApplyTo fills up NodeLifecycleController config with options.
func (o *NodeLifecycleControllerOptions) ApplyTo(cfg *nodelifecycleconfig.NodeLifecycleControllerConfiguration) error {
	if o == nil {
		return nil
	}

	cfg.EnableTaintManager = o.EnableTaintManager
	cfg.NodeStartupGracePeriod = o.NodeStartupGracePeriod
	cfg.NodeMonitorGracePeriod = o.NodeMonitorGracePeriod
	cfg.PodEvictionTimeout = o.PodEvictionTimeout
	cfg.NodeEvictionRate = o.NodeEvictionRate
	cfg.SecondaryNodeEvictionRate = o.SecondaryNodeEvictionRate
	cfg.LargeClusterSizeThreshold = o.LargeClusterSizeThreshold
	cfg.UnhealthyZoneThreshold = o.UnhealthyZoneThreshold

	return nil
}

// Validate checks validation of NodeLifecycleControllerOptions.
func (o *NodeLifecycleControllerOptions) Validate() []error {
	if o == nil {
		return nil
	}

	errs := []error{}
	return errs
}
