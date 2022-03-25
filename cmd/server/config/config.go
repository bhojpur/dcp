package config

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
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/server/options"
	dcpcorev1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	dcpclientset "github.com/bhojpur/dcp/pkg/appmanager/client/clientset/versioned"
	dcpinformers "github.com/bhojpur/dcp/pkg/appmanager/client/informers/externalversions"
	dcpv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/client/informers/externalversions/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/engine/cachemanager"
	"github.com/bhojpur/dcp/pkg/engine/filter"
	"github.com/bhojpur/dcp/pkg/engine/filter/discardcloudservice"
	"github.com/bhojpur/dcp/pkg/engine/filter/ingresscontroller"
	"github.com/bhojpur/dcp/pkg/engine/filter/initializer"
	"github.com/bhojpur/dcp/pkg/engine/filter/masterservice"
	"github.com/bhojpur/dcp/pkg/engine/filter/servicetopology"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/meta"
	"github.com/bhojpur/dcp/pkg/engine/kubernetes/serializer"
	"github.com/bhojpur/dcp/pkg/engine/storage/factory"
	"github.com/bhojpur/dcp/pkg/engine/util"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

// EngineConfiguration represents configuration of Bhojpur DCP engine
type EngineConfiguration struct {
	LBMode                           string
	RemoteServers                    []*url.URL
	EngineServerAddr                 string
	EngineCertOrganizations          []string
	EngineProxyServerAddr            string
	EngineProxyServerSecureAddr      string
	EngineProxyServerDummyAddr       string
	EngineProxyServerSecureDummyAddr string
	GCFrequency                      int
	CertMgrMode                      string
	KubeletRootCAFilePath            string
	KubeletPairFilePath              string
	NodeName                         string
	HeartbeatFailedRetry             int
	HeartbeatHealthyThreshold        int
	HeartbeatTimeoutSeconds          int
	MaxRequestInFlight               int
	JoinToken                        string
	RootDir                          string
	EnableProfiling                  bool
	EnableDummyIf                    bool
	EnableIptables                   bool
	HubAgentDummyIfName              string
	StorageWrapper                   cachemanager.StorageWrapper
	SerializerManager                *serializer.SerializerManager
	RESTMapperManager                *meta.RESTMapperManager
	TLSConfig                        *tls.Config
	SharedFactory                    informers.SharedInformerFactory
	DcpSharedFactory                 dcpinformers.SharedInformerFactory
	WorkingMode                      util.WorkingMode
	KubeletHealthGracePeriod         time.Duration
	FilterChain                      filter.Interface
}

// Complete converts *options.EngineOptions to *EngineConfiguration
func Complete(options *options.EngineOptions) (*EngineConfiguration, error) {
	us, err := parseRemoteServers(options.ServerAddr)
	if err != nil {
		return nil, err
	}

	hubCertOrgs := make([]string, 0)
	if options.EngineCertOrganizations != "" {
		for _, orgStr := range strings.Split(options.EngineCertOrganizations, ",") {
			hubCertOrgs = append(hubCertOrgs, orgStr)
		}
	}

	storageManager, err := factory.CreateStorage(options.DiskCachePath)
	if err != nil {
		klog.Errorf("could not create storage manager, %v", err)
		return nil, err
	}
	storageWrapper := cachemanager.NewStorageWrapper(storageManager)
	serializerManager := serializer.NewSerializerManager()
	restMapperManager := meta.NewRESTMapperManager(storageManager)

	hubServerAddr := net.JoinHostPort(options.EngineHost, options.EnginePort)
	proxyServerAddr := net.JoinHostPort(options.EngineHost, options.EngineProxyPort)
	proxySecureServerAddr := net.JoinHostPort(options.EngineHost, options.EngineProxySecurePort)
	proxyServerDummyAddr := net.JoinHostPort(options.HubAgentDummyIfIP, options.EngineProxyPort)
	proxySecureServerDummyAddr := net.JoinHostPort(options.HubAgentDummyIfIP, options.EngineProxySecurePort)
	workingMode := util.WorkingMode(options.WorkingMode)

	var filterChain filter.Interface
	var filters *filter.Filters
	var serviceTopologyFilterEnabled bool
	var mutatedMasterServiceAddr string
	if options.EnableResourceFilter {
		if options.WorkingMode == string(util.WorkingModeCloud) {
			options.DisabledResourceFilters = append(options.DisabledResourceFilters, filter.DisabledInCloudMode...)
		}
		filters = filter.NewFilters(options.DisabledResourceFilters)
		registerAllFilters(filters)

		serviceTopologyFilterEnabled = filters.Enabled(filter.ServiceTopologyFilterName)
		mutatedMasterServiceAddr = us[0].Host
		if options.AccessServerThroughHub {
			if options.EnableDummyIf {
				mutatedMasterServiceAddr = proxySecureServerDummyAddr
			} else {
				mutatedMasterServiceAddr = proxySecureServerAddr
			}
		}
	}

	sharedFactory, dcpSharedFactory, err := createSharedInformers(fmt.Sprintf("http://%s", proxyServerAddr))
	if err != nil {
		return nil, err
	}
	registerInformers(sharedFactory, dcpSharedFactory, workingMode, serviceTopologyFilterEnabled, options.NodePoolName, options.NodeName)
	filterChain, err = createFilterChain(filters, sharedFactory, dcpSharedFactory, serializerManager, storageWrapper, workingMode, options.NodeName, mutatedMasterServiceAddr)
	if err != nil {
		return nil, err
	}

	cfg := &EngineConfiguration{
		LBMode:                           options.LBMode,
		RemoteServers:                    us,
		EngineServerAddr:                 hubServerAddr,
		EngineCertOrganizations:          hubCertOrgs,
		EngineProxyServerAddr:            proxyServerAddr,
		EngineProxyServerSecureAddr:      proxySecureServerAddr,
		EngineProxyServerDummyAddr:       proxyServerDummyAddr,
		EngineProxyServerSecureDummyAddr: proxySecureServerDummyAddr,
		GCFrequency:                      options.GCFrequency,
		CertMgrMode:                      options.CertMgrMode,
		KubeletRootCAFilePath:            options.KubeletRootCAFilePath,
		KubeletPairFilePath:              options.KubeletPairFilePath,
		NodeName:                         options.NodeName,
		HeartbeatFailedRetry:             options.HeartbeatFailedRetry,
		HeartbeatHealthyThreshold:        options.HeartbeatHealthyThreshold,
		HeartbeatTimeoutSeconds:          options.HeartbeatTimeoutSeconds,
		MaxRequestInFlight:               options.MaxRequestInFlight,
		JoinToken:                        options.JoinToken,
		RootDir:                          options.RootDir,
		EnableProfiling:                  options.EnableProfiling,
		EnableDummyIf:                    options.EnableDummyIf,
		EnableIptables:                   options.EnableIptables,
		HubAgentDummyIfName:              options.HubAgentDummyIfName,
		WorkingMode:                      workingMode,
		StorageWrapper:                   storageWrapper,
		SerializerManager:                serializerManager,
		RESTMapperManager:                restMapperManager,
		SharedFactory:                    sharedFactory,
		DcpSharedFactory:                 dcpSharedFactory,
		KubeletHealthGracePeriod:         options.KubeletHealthGracePeriod,
		FilterChain:                      filterChain,
	}

	return cfg, nil
}

func parseRemoteServers(serverAddr string) ([]*url.URL, error) {
	if serverAddr == "" {
		return make([]*url.URL, 0), fmt.Errorf("--server-addr should be set for hub agent")
	}
	servers := strings.Split(serverAddr, ",")
	us := make([]*url.URL, 0, len(servers))
	remoteServers := make([]string, 0, len(servers))
	for _, server := range servers {
		u, err := url.Parse(server)
		if err != nil {
			klog.Errorf("failed to parse server address %s, %v", servers, err)
			return us, err
		}
		if u.Scheme == "" {
			u.Scheme = "https"
		} else if u.Scheme != "https" {
			return us, fmt.Errorf("only https scheme is supported for server address(%s)", serverAddr)
		}
		us = append(us, u)
		remoteServers = append(remoteServers, u.String())
	}

	if len(us) < 1 {
		return us, fmt.Errorf("no server address is set, can not connect remote server")
	}
	klog.Infof("%s would connect remote servers: %s", projectinfo.GetEngineName(), strings.Join(remoteServers, ","))

	return us, nil
}

// createSharedInformers create sharedInformers from the given proxyAddr.
func createSharedInformers(proxyAddr string) (informers.SharedInformerFactory, dcpinformers.SharedInformerFactory, error) {
	var kubeConfig *rest.Config
	var err error
	kubeConfig, err = clientcmd.BuildConfigFromFlags(proxyAddr, "")
	if err != nil {
		return nil, nil, err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	dcpClient, err := dcpclientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	return informers.NewSharedInformerFactory(client, 24*time.Hour),
		dcpinformers.NewSharedInformerFactory(dcpClient, 24*time.Hour), nil
}

// registerInformers reconstruct node/nodePool/configmap informers
func registerInformers(informerFactory informers.SharedInformerFactory,
	dcpInformerFactory dcpinformers.SharedInformerFactory,
	workingMode util.WorkingMode,
	serviceTopologyFilterEnabled bool,
	nodePoolName, nodeName string) {
	// skip construct node/nodePool informers if service topology filter disabled
	if serviceTopologyFilterEnabled {
		if workingMode == util.WorkingModeCloud {
			newNodeInformer := func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
				tweakListOptions := func(options *metav1.ListOptions) {
					options.FieldSelector = fields.Set{"metadata.name": nodeName}.String()
				}
				return coreinformers.NewFilteredNodeInformer(client, resyncPeriod, nil, tweakListOptions)
			}
			informerFactory.InformerFor(&corev1.Node{}, newNodeInformer)
		}

		if len(nodePoolName) != 0 {
			newNodePoolInformer := func(client dcpclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
				tweakListOptions := func(options *metav1.ListOptions) {
					options.FieldSelector = fields.Set{"metadata.name": nodePoolName}.String()
				}
				return dcpv1alpha1.NewFilteredNodePoolInformer(client, resyncPeriod, nil, tweakListOptions)
			}

			dcpInformerFactory.InformerFor(&dcpcorev1alpha1.NodePool{}, newNodePoolInformer)
		}
	}

	if workingMode == util.WorkingModeEdge {
		newConfigmapInformer := func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
			tweakListOptions := func(options *metav1.ListOptions) {
				options.FieldSelector = fields.Set{"metadata.name": util.EngineConfigMapName}.String()
			}
			return coreinformers.NewFilteredConfigMapInformer(client, util.EngineNamespace, resyncPeriod, nil, tweakListOptions)
		}
		informerFactory.InformerFor(&corev1.ConfigMap{}, newConfigmapInformer)
	}
}

// registerAllFilters by order, the front registered filter will be
// called before the behind registered ones.
func registerAllFilters(filters *filter.Filters) {
	servicetopology.Register(filters)
	masterservice.Register(filters)
	discardcloudservice.Register(filters)
	ingresscontroller.Register(filters)
}

// createFilterChain return union filters that initializations completed.
func createFilterChain(filters *filter.Filters,
	sharedFactory informers.SharedInformerFactory,
	dcpSharedFactory dcpinformers.SharedInformerFactory,
	serializerManager *serializer.SerializerManager,
	storageWrapper cachemanager.StorageWrapper,
	workingMode util.WorkingMode,
	nodeName, mutatedMasterServiceAddr string) (filter.Interface, error) {
	if filters == nil {
		return nil, nil
	}

	genericInitializer := initializer.New(sharedFactory, dcpSharedFactory, serializerManager, storageWrapper, nodeName, mutatedMasterServiceAddr, workingMode)
	initializerChain := filter.FilterInitializers{}
	initializerChain = append(initializerChain, genericInitializer)
	return filters.NewFromFilters(initializerChain)
}
