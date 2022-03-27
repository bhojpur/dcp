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

// It provides the flags used for the controller manager.

import (
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	componentbaseconfig "k8s.io/component-base/config"
	cmconfig "k8s.io/controller-manager/config"
	cmoptions "k8s.io/controller-manager/options"
	"k8s.io/klog/v2"
	nodelifecycleconfig "k8s.io/kube-controller-manager/config/v1alpha1"
	utilpointer "k8s.io/utils/pointer"

	dcpcontrollerconfig "github.com/bhojpur/dcp/cmd/grid/controller-manager/config"
	"github.com/bhojpur/dcp/pkg/projectinfo"
)

const (
	// DcpControllerManagerUserAgent is the userAgent name when starting controller managers.
	DcpControllerManagerUserAgent = "controller-manager"
)

// DcpControllerManagerOptions is the main context object for the kube-controller manager.
type DcpControllerManagerOptions struct {
	Generic                 *cmoptions.GenericControllerManagerConfigurationOptions
	NodeLifecycleController *NodeLifecycleControllerOptions
	Master                  string
	Kubeconfig              string
	Version                 bool
}

// NewDcpControllerManagerOptions creates a new DcpControllerManagerOptions with a default config.
func NewDcpControllerManagerOptions() (*DcpControllerManagerOptions, error) {
	generic := cmconfig.GenericControllerManagerConfiguration{
		Address:                 "0.0.0.0",
		Port:                    10266,
		MinResyncPeriod:         metav1.Duration{Duration: 12 * time.Hour},
		ControllerStartInterval: metav1.Duration{Duration: 0 * time.Second},
		Controllers:             []string{"*"},
		ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
			ContentType: "application/vnd.kubernetes.protobuf",
			QPS:         50.0,
			Burst:       100,
		},
		LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
			LeaseDuration: metav1.Duration{Duration: 15 * time.Second},
			RenewDeadline: metav1.Duration{Duration: 10 * time.Second},
			RetryPeriod:   metav1.Duration{Duration: 2 * time.Second},
			ResourceLock:  resourcelock.LeasesResourceLock,
			LeaderElect:   true,
		},
	}

	s := DcpControllerManagerOptions{
		Generic: cmoptions.NewGenericControllerManagerConfigurationOptions(&generic),
		NodeLifecycleController: &NodeLifecycleControllerOptions{
			NodeLifecycleControllerConfiguration: &nodelifecycleconfig.NodeLifecycleControllerConfiguration{
				EnableTaintManager:     utilpointer.BoolPtr(true),
				PodEvictionTimeout:     metav1.Duration{Duration: 5 * time.Minute},
				NodeMonitorGracePeriod: metav1.Duration{Duration: 40 * time.Second},
				NodeStartupGracePeriod: metav1.Duration{Duration: 60 * time.Second},
			},
		},
	}

	return &s, nil
}

// Flags returns flags for a specific APIServer by section name
func (s *DcpControllerManagerOptions) Flags(allControllers []string, disabledByDefaultControllers []string) cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	s.Generic.AddFlags(&fss, allControllers, disabledByDefaultControllers)
	s.NodeLifecycleController.AddFlags(fss.FlagSet("nodelifecycle controller"))

	fs := fss.FlagSet("misc")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	utilfeature.DefaultMutableFeatureGate.AddFlag(fss.FlagSet("generic"))
	fs.BoolVar(&s.Version, "version", s.Version, "print the version information.")

	return fss
}

// ApplyTo fills up controller manager config with options.
func (s *DcpControllerManagerOptions) ApplyTo(c *dcpcontrollerconfig.Config) error {
	if err := s.Generic.ApplyTo(&c.ComponentConfig.Generic); err != nil {
		return err
	}

	if err := s.NodeLifecycleController.ApplyTo(&c.ComponentConfig.NodeLifecycleController); err != nil {
		return err
	}

	return nil
}

// Validate is used to validate the options and config before launching the controller manager
func (s *DcpControllerManagerOptions) Validate(allControllers []string, disabledByDefaultControllers []string) error {
	var errs []error

	errs = append(errs, s.Generic.Validate(allControllers, disabledByDefaultControllers)...)
	errs = append(errs, s.NodeLifecycleController.Validate()...)

	// TODO: validate component config, master and kubeconfig

	return utilerrors.NewAggregate(errs)
}

// Config return a controller manager config objective
func (s DcpControllerManagerOptions) Config(allControllers []string, disabledByDefaultControllers []string) (*dcpcontrollerconfig.Config, error) {
	if err := s.Validate(allControllers, disabledByDefaultControllers); err != nil {
		return nil, err
	}

	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeconfig.DisableCompression = true
	kubeconfig.ContentConfig.AcceptContentTypes = s.Generic.ClientConnection.AcceptContentTypes
	kubeconfig.ContentConfig.ContentType = s.Generic.ClientConnection.ContentType
	kubeconfig.QPS = s.Generic.ClientConnection.QPS
	kubeconfig.Burst = int(s.Generic.ClientConnection.Burst)

	client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, projectinfo.GetControllerManagerName()))
	if err != nil {
		return nil, err
	}

	// shallow copy, do not modify the kubeconfig.Timeout.
	config := *kubeconfig
	config.Timeout = s.Generic.LeaderElection.RenewDeadline.Duration
	leaderElectionClient := clientset.NewForConfigOrDie(restclient.AddUserAgent(&config, "leader-election"))

	eventRecorder := createRecorder(client, projectinfo.GetControllerManagerName())

	c := &dcpcontrollerconfig.Config{
		Client:               client,
		Kubeconfig:           kubeconfig,
		EventRecorder:        eventRecorder,
		LeaderElectionClient: leaderElectionClient,
	}
	if err := s.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}
