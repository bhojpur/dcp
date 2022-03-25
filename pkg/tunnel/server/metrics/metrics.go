package metrics

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
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/bhojpur/dcp/pkg/projectinfo"
)

var (
	namespace = strings.ReplaceAll(projectinfo.GetTunnelName(), "-", "_")
	subsystem = "server"
)

var (
	// Metrics provides access to all tunnel server metrics.
	Metrics = newTunnelServerMetrics()
)

type TunnelServerMetrics struct {
	proxyingRequestsCollector *prometheus.GaugeVec
	proxyingRequestsGauge     prometheus.Gauge
	cloudNodeGauge            prometheus.Gauge
}

func newTunnelServerMetrics() *TunnelServerMetrics {
	proxyingRequestsCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "in_proxy_requests",
			Help:      "how many http requests are proxying by tunnel server",
		},
		[]string{"verb", "path"})
	proxyingRequestsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "total_in_proxy_requests",
			Help:      "the number of http requests are proxying by tunnel server",
		})
	cloudNodeGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cloud_nodes_counter",
			Help:      "counter of cloud nodes that do not run tunnel agent",
		})

	prometheus.MustRegister(proxyingRequestsCollector)
	prometheus.MustRegister(proxyingRequestsGauge)
	prometheus.MustRegister(cloudNodeGauge)
	return &TunnelServerMetrics{
		proxyingRequestsCollector: proxyingRequestsCollector,
		proxyingRequestsGauge:     proxyingRequestsGauge,
		cloudNodeGauge:            cloudNodeGauge,
	}
}

func (tsm *TunnelServerMetrics) Reset() {
	tsm.proxyingRequestsCollector.Reset()
	tsm.proxyingRequestsGauge.Set(float64(0))
	tsm.cloudNodeGauge.Set(float64(0))
}

func (tsm *TunnelServerMetrics) IncInFlightRequests(verb, path string) {
	tsm.proxyingRequestsCollector.WithLabelValues(verb, path).Inc()
	tsm.proxyingRequestsGauge.Inc()
}

func (tsm *TunnelServerMetrics) DecInFlightRequests(verb, path string) {
	tsm.proxyingRequestsCollector.WithLabelValues(verb, path).Dec()
	tsm.proxyingRequestsGauge.Dec()
}

func (tsm *TunnelServerMetrics) ObserveCloudNodes(cnt int) {
	tsm.cloudNodeGauge.Set(float64(cnt))
}
