package dcpinit

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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/pkg/client/util"
	"github.com/bhojpur/dcp/pkg/client/util/edgenode"
	strutil "github.com/bhojpur/dcp/pkg/client/util/strings"
	tmplutil "github.com/bhojpur/dcp/pkg/client/util/templates"
)

const (
	// APIServerAdvertiseAddress flag sets the IP address the API Server will advertise it's listening on. Specify '0.0.0.0' to use the address of the default network interface.
	APIServerAdvertiseAddress = "apiserver-advertise-address"
	//TunnelServerAddress flag sets the IP address of Tunnel Server.
	TunnelServerAddress = "tunnel-server-address"
	// NetworkingServiceSubnet flag sets the range of IP address for service VIPs.
	NetworkingServiceSubnet = "service-cidr"
	// NetworkingPodSubnet flag sets the range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node.
	NetworkingPodSubnet = "pod-network-cidr"
	// BhojpurDcpVersion flag sets the Bhojpur DCP version for the control plane.
	BhojpurDcpVersion = "bhojpur-dcp-version"
	// ImageRepository flag sets the container registry to pull control plane images from.
	ImageRepository = "image-repository"
	// PassWd flag is the password of master server.
	PassWd = "passwd"

	TmpDownloadDir = "/tmp"

	SealerUrlFormat      = "https://github.com/alibaba/sealer/releases/download/%s/sealer-%s-linux-%s.tar.gz"
	DefaultSealerVersion = "v0.6.1"

	InitClusterImage = "%s/bhojpur-dcp-cluster:%s"
	SealerRunCmd     = "sealer apply -f %s/Clusterfile"

	fileMode = 0666
	dirMode  = 0755

	BhojpurClusterfile = `
apiVersion: sealer.cloud/v2
kind: Cluster
metadata:
  name: my-cluster
spec:
  hosts:
  - ips:
    - {{.apiserver_address}}
    roles:
    - master
  image: {{.cluster_image}}
  ssh:
    passwd: {{.passwd}}
    pk: /root/.ssh/id_rsa
    user: root
  env:
  - TunnelServerAddress={{.tunnel_server_address}}
---
apiVersion: sealer.cloud/v2
kind: KubeadmConfig
metadata:
  name: default-kubernetes-config
spec:
  networking:
    {{if .pod_subnet }}
    podSubnet: {{.pod_subnet}}
    {{end}}
    {{if .service_subnet}}
    serviceSubnet: {{.service_subnet}}
    {{end}}
  controllerManager:
    extraArgs:
      controllers: -nodelifecycle,*,bootstrapsigner,tokencleaner
`
)

var (
	ValidSealerVersions = []string{
		"v0.6.1",
	}
)

// clusterInitializer init a node to master of Bhojpur DCP cluster
type clusterInitializer struct {
	InitOptions
}

// NewCmdInit use tool sealer to initializer a master of Bhojpur DCP cluster.
// It will deploy all Bhojpur DCP components, such as app-manager, tunnel-server, etc.
func NewCmdInit() *cobra.Command {
	o := NewInitOptions()

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Run this command in order to set up the Bhojpur DCP control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(); err != nil {
				return err
			}
			initializer := NewInitializerWithOptions(o)
			if err := initializer.Run(); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}

	addFlags(cmd.Flags(), o)
	return cmd
}

func addFlags(flagset *flag.FlagSet, o *InitOptions) {
	flagset.StringVarP(
		&o.AdvertiseAddress, APIServerAdvertiseAddress, "", o.AdvertiseAddress,
		"The IP address the API Server will advertise it's listening on.",
	)
	flagset.StringVarP(
		&o.TunnelServerAddress, TunnelServerAddress, "", o.TunnelServerAddress,
		"The tunnel-server address.")
	flagset.StringVarP(
		&o.ServiceSubnet, NetworkingServiceSubnet, "", o.ServiceSubnet,
		"Use alternative range of IP address for service VIPs.",
	)
	flagset.StringVarP(
		&o.PodSubnet, NetworkingPodSubnet, "", o.PodSubnet,
		"Specify range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node.",
	)
	flagset.StringVarP(&o.Password, PassWd, "p", o.Password,
		"set master server ssh password",
	)
	flagset.StringVarP(
		&o.BhojpurDcpVersion, BhojpurDcpVersion, "", o.BhojpurDcpVersion,
		`Choose a specific Bhojpur DCP version for the control plane.`,
	)
	flagset.StringVarP(&o.ImageRepository, ImageRepository, "", o.ImageRepository,
		"Choose a registry to pull cluster images from",
	)
}

func NewInitializerWithOptions(o *InitOptions) *clusterInitializer {
	return &clusterInitializer{
		*o,
	}
}

// Run use sealer to initialize the master node.
func (ci *clusterInitializer) Run() error {
	if err := CheckAndInstallSealer(); err != nil {
		return err
	}

	if err := ci.PrepareClusterfile(); err != nil {
		return err
	}

	if err := ci.InstallCluster(); err != nil {
		return err
	}
	return nil
}

// CheckAndInstallSealer install sealer, skip install if it exists
func CheckAndInstallSealer() error {
	klog.Infof("Check and install sealer")
	sealerExist := false
	if _, err := exec.LookPath("sealer"); err == nil {
		if b, err := exec.Command("sealer", "version").CombinedOutput(); err == nil {
			info := make(map[string]string)
			if err := json.Unmarshal(b, &info); err != nil {
				return fmt.Errorf("Can't get the existing sealer version: %v", err)
			}
			sealerVersion := info["gitVersion"]
			if strutil.IsInStringLst(ValidSealerVersions, sealerVersion) {
				klog.Infof("Sealer %s already exist, skip install.", sealerVersion)
				sealerExist = true
			} else {
				return fmt.Errorf("The existing sealer version %s is not supported, please clean it. Valid server versions are %v.", sealerVersion, ValidSealerVersions)
			}
		}
	}

	if !sealerExist {
		// download and install sealer
		packageUrl := fmt.Sprintf(SealerUrlFormat, DefaultSealerVersion, DefaultSealerVersion, runtime.GOARCH)
		savePath := fmt.Sprintf("%s/sealer-%s-linux-%s.tar.gz", TmpDownloadDir, DefaultSealerVersion, runtime.GOARCH)
		klog.V(1).Infof("Download sealer from: %s", packageUrl)
		if err := util.DownloadFile(packageUrl, savePath, 3); err != nil {
			return fmt.Errorf("Download sealer fail: %v", err)
		}
		if err := util.Untar(savePath, TmpDownloadDir); err != nil {
			return err
		}
		comp := "sealer"
		target := fmt.Sprintf("/usr/bin/%s", comp)
		if err := edgenode.CopyFile(TmpDownloadDir+"/"+comp, target, dirMode); err != nil {
			return err
		}
	}
	return nil
}

// InstallCluster initialize the master of Bhojpur DCP cluster by calling sealer
func (ci *clusterInitializer) InstallCluster() error {
	klog.Infof("init a Bhojpur DCP cluster")
	runCmd := fmt.Sprintf(SealerRunCmd, TmpDownloadDir)
	cmd := exec.Command("bash", "-c", runCmd)
	return execCmd(cmd)
}

// PrepareClusterfile fill the template and write the Clusterfile to the /tmp
func (ci *clusterInitializer) PrepareClusterfile() error {
	klog.Infof("generate Clusterfile for Bhojpur DCP")
	err := os.MkdirAll(TmpDownloadDir, dirMode)
	if err != nil {
		return err
	}

	clusterfile, err := tmplutil.SubsituteTemplate(BhojpurClusterfile, map[string]string{
		"apiserver_address":     ci.AdvertiseAddress,
		"cluster_image":         fmt.Sprintf(InitClusterImage, ci.ImageRepository, ci.BhojpurDcpVersion),
		"passwd":                ci.Password,
		"pod_subnet":            ci.PodSubnet,
		"service_subnet":        ci.ServiceSubnet,
		"tunnel_server_address": ci.TunnelServerAddress,
	})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/Clusterfile", TmpDownloadDir), []byte(clusterfile), fileMode)
	if err != nil {
		return err
	}
	return nil
}

func execCmd(cmd *exec.Cmd) error {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf(outStr)
	if err != nil {
		pos := strings.Index(errStr, "Usage:")
		fmt.Printf(errStr[:pos])
	}
	return err
}
