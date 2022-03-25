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

	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	namespace = "node"
	subsystem = strings.ReplaceAll(projectinfo.GetEngineName(), "-", "_")
)

var (
	// Metrics provides access to all hub agent metrics.
	Metrics = newHubMetrics()
)

type HubMetrics struct {
	serversHealthyCollector   *prometheus.GaugeVec
	inFlightRequestsCollector *prometheus.GaugeVec
	inFlightRequestsGauge     prometheus.Gauge
	rejectedRequestsCounter   prometheus.Counter
	closableConnsCollector    *prometheus.GaugeVec
	proxyTrafficCollector     *prometheus.CounterVec
}

func newHubMetrics() *HubMetrics {
	serversHealthyCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "server_healthy_status",
			Help:      "healthy status of remote servers. 1: healthy, 0: unhealthy",
		},
		[]string{"server"})
	inFlightRequestsCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "in_flight_requests_collector",
			Help:      "collector of in flight requests handling by hub agent",
		},
		[]string{"verb", "resource", "subresources", "client"})
	inFlightRequestsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "in_flight_requests_total",
			Help:      "total of in flight requests handling by hub agent",
		})
	rejectedRequestsCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "rejected_requests_counter",
			Help:      "counter of rejected requests for exceeding in flight limit in hub agent",
		})
	closableConnsCollector := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "closable_conns_collector",
			Help:      "collector of underlay tcp connection from hub agent to remote server",
		},
		[]string{"server"})
	proxyTrafficCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "proxy_traffic_collector",
			Help:      "collector of proxy response traffic by hub agent(unit: byte)",
		},
		[]string{"client", "verb", "resource", "subresources"})
	prometheus.MustRegister(serversHealthyCollector)
	prometheus.MustRegister(inFlightRequestsCollector)
	prometheus.MustRegister(inFlightRequestsGauge)
	prometheus.MustRegister(rejectedRequestsCounter)
	prometheus.MustRegister(closableConnsCollector)
	prometheus.MustRegister(proxyTrafficCollector)
	return &HubMetrics{
		serversHealthyCollector:   serversHealthyCollector,
		inFlightRequestsCollector: inFlightRequestsCollector,
		inFlightRequestsGauge:     inFlightRequestsGauge,
		rejectedRequestsCounter:   rejectedRequestsCounter,
		closableConnsCollector:    closableConnsCollector,
		proxyTrafficCollector:     proxyTrafficCollector,
	}
}

func (hm *HubMetrics) Reset() {
	hm.serversHealthyCollector.Reset()
	hm.inFlightRequestsCollector.Reset()
	hm.inFlightRequestsGauge.Set(float64(0))
	hm.closableConnsCollector.Reset()
	hm.proxyTrafficCollector.Reset()
}

func (hm *HubMetrics) ObserveServerHealthy(server string, status int) {
	hm.serversHealthyCollector.WithLabelValues(server).Set(float64(status))
}

func (hm *HubMetrics) IncInFlightRequests(verb, resource, subresource, client string) {
	hm.inFlightRequestsCollector.WithLabelValues(verb, resource, subresource, client).Inc()
	hm.inFlightRequestsGauge.Inc()
}

func (hm *HubMetrics) DecInFlightRequests(verb, resource, subresource, client string) {
	hm.inFlightRequestsCollector.WithLabelValues(verb, resource, subresource, client).Dec()
	hm.inFlightRequestsGauge.Dec()
}

func (hm *HubMetrics) IncRejectedRequestCounter() {
	hm.rejectedRequestsCounter.Inc()
}

func (hm *HubMetrics) IncClosableConns(server string) {
	hm.closableConnsCollector.WithLabelValues(server).Inc()
}

func (hm *HubMetrics) DecClosableConns(server string) {
	hm.closableConnsCollector.WithLabelValues(server).Dec()
}

func (hm *HubMetrics) SetClosableConns(server string, cnt int) {
	hm.closableConnsCollector.WithLabelValues(server).Set(float64(cnt))
}

func (hm *HubMetrics) AddProxyTrafficCollector(client, verb, resource, subresource string, size int) {
	if size > 0 {
		hm.proxyTrafficCollector.WithLabelValues(client, verb, resource, subresource).Add(float64(size))
	}
}
