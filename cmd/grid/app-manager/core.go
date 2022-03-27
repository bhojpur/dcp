package app

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
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/bhojpur/dcp/cmd/grid/app-manager/options"
	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	extclient "github.com/bhojpur/dcp/pkg/appmanager/client"
	"github.com/bhojpur/dcp/pkg/appmanager/constant"
	"github.com/bhojpur/dcp/pkg/appmanager/controller"
	"github.com/bhojpur/dcp/pkg/appmanager/util/fieldindex"
	"github.com/bhojpur/dcp/pkg/appmanager/webhook"
	"github.com/bhojpur/dcp/pkg/projectinfo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	restConfigQPS   = flag.Int("rest-config-qps", 30, "QPS of rest config.")
	restConfigBurst = flag.Int("rest-config-burst", 50, "Burst of rest config.")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1alpha1.AddToScheme(clientgoscheme.Scheme)

	_ = appsv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// NewCmdAppManager creates a *cobra.Command object with default parameters
func NewCmdAppManager(stopCh <-chan struct{}) *cobra.Command {
	dcpAppOptions := options.NewAppOptions()

	cmd := &cobra.Command{
		Use:   projectinfo.GetAppManagerName(),
		Short: "Launch Bhojpur DCP " + projectinfo.GetAppManagerName(),
		Long:  "Launch Bhojpur DCP " + projectinfo.GetAppManagerName(),
		Run: func(cmd *cobra.Command, args []string) {
			if dcpAppOptions.Version {
				fmt.Printf("%s: %#v\n", projectinfo.GetAppManagerName(), projectinfo.Get())
				return
			}

			fmt.Printf("%s version: %#v\n", projectinfo.GetAppManagerName(), projectinfo.Get())

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if err := options.ValidateOptions(dcpAppOptions); err != nil {
				klog.Fatalf("validate options: %v", err)
			}

			Run(dcpAppOptions)
		},
	}

	dcpAppOptions.AddFlags(cmd.Flags())
	return cmd
}

func Run(opts *options.AppOptions) {
	if opts.EnablePprof {
		go func() {
			if err := http.ListenAndServe(opts.PprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof")
			}
		}()
	}

	//ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	ctrl.SetLogger(klogr.New())

	cfg := ctrl.GetConfigOrDie()
	setRestConfig(cfg)

	cacheDisableObjs := []client.Object{
		&appsv1alpha1.DcpIngress{},
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         opts.MetricsAddr,
		HealthProbeBindAddress:     opts.HealthProbeAddr,
		LeaderElection:             opts.EnableLeaderElection,
		LeaderElectionID:           "app-manager",
		LeaderElectionNamespace:    opts.LeaderElectionNamespace,
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock, // use lease to election
		Namespace:                  opts.Namespace,
		ClientDisableCacheFor:      cacheDisableObjs,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	setupLog.Info("register field index")
	if err := fieldindex.RegisterFieldIndexes(mgr.GetCache()); err != nil {
		setupLog.Error(err, "failed to register field index")
		os.Exit(1)
	}

	setupLog.Info("new clientset registry")
	err = extclient.NewRegistry(mgr)
	if err != nil {
		setupLog.Error(err, "unable to init Bhojpur DCP application clientset and informer")
		os.Exit(1)
	}

	setupLog.Info("setup controllers")

	ctx := genOptCtx(opts.CreateDefaultPool)
	if err = controller.SetupWithManager(mgr, ctx); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		os.Exit(1)
	}

	setupLog.Info("setup webhook")
	if err = webhook.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup webhook")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	stopCh := ctrl.SetupSignalHandler()
	setupLog.Info("initialize webhook")
	if err := webhook.Initialize(mgr, stopCh.Done()); err != nil {
		setupLog.Error(err, "unable to initialize webhook")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("webhook-ready", webhook.Checker); err != nil {
		setupLog.Error(err, "unable to add readyz check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}

func genOptCtx(createDefaultPool bool) context.Context {
	return context.WithValue(context.Background(),
		constant.ContextKeyCreateDefaultPool, createDefaultPool)
}

func setRestConfig(c *rest.Config) {
	if *restConfigQPS > 0 {
		c.QPS = float32(*restConfigQPS)
	}
	if *restConfigBurst > 0 {
		c.Burst = *restConfigBurst
	}
}
