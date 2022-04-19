package phases

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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"

	"github.com/bhojpur/dcp/cmd/grid/dcpctl/join/joindata"
	dcpconstants "github.com/bhojpur/dcp/pkg/client/constants"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	kubeutil "github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/phases/kubelet"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/apiclient"
	kubeletconfig "github.com/bhojpur/dcp/pkg/client/kubernetes/kubelet/apis/config"
	kubeletscheme "github.com/bhojpur/dcp/pkg/client/kubernetes/kubelet/apis/config/scheme"
	kubeletcodec "github.com/bhojpur/dcp/pkg/client/kubernetes/kubelet/kubeletconfig/util/codec"
	"github.com/bhojpur/dcp/pkg/client/util/edgenode"
	"github.com/bhojpur/dcp/pkg/client/util/templates"
)

// NewEdgeNodePhase creates a client workflow phase that start kubelet on a edge node.
func NewEdgeNodePhase() workflow.Phase {
	return workflow.Phase{
		Name:  "Join node to Bhojpur DCP cluster.",
		Short: "Join node",
		Run:   runJoinNode,
		InheritFlags: []string{
			options.TokenStr,
			options.NodeCRISocket,
			options.NodeName,
			options.IgnorePreflightErrors,
		},
	}
}

// runJoinNode executes the node join process.
func runJoinNode(c workflow.RunData) error {
	data, ok := c.(joindata.DcpJoinData)
	if !ok {
		return fmt.Errorf("Join edge-node phase invoked with an invalid data struct. ")
	}

	err := writeKubeletConfigFile(data.BootstrapClient(), data)
	if err != nil {
		return err
	}
	if err := addEngineStaticYaml(data, filepath.Join(constants.KubernetesDir, constants.ManifestsSubDirName)); err != nil {
		return err
	}
	klog.Info("[kubelet-start] Starting the kubelet")
	kubeutil.TryStartKubelet()
	return nil
}

// writeKubeletConfigFile write kubelet configuration into local disk.
func writeKubeletConfigFile(bootstrapClient *clientset.Clientset, data joindata.DcpJoinData) error {
	kubeletVersion, err := version.ParseSemantic(data.KubernetesVersion())
	if err != nil {
		return err
	}

	// Write the configuration for the kubelet (using the bootstrap token credentials) to disk so the kubelet can start
	_, err = downloadConfig(bootstrapClient, kubeletVersion, constants.KubeletRunDirectory)
	if err != nil {
		return err
	}

	if err := kubeutil.WriteKubeletDynamicEnvFile(data, constants.KubeletRunDirectory); err != nil {
		return err
	}
	return nil
}

// downloadConfig downloads the kubelet configuration from a ConfigMap and writes it to disk.
// Used at "kubeadm join" time
func downloadConfig(client clientset.Interface, kubeletVersion *version.Version, kubeletDir string) (*kubeletconfig.KubeletConfiguration, error) {
	// Download the ConfigMap from the cluster based on what version the kubelet is
	configMapName := constants.GetKubeletConfigMapName(kubeletVersion)

	klog.Infof("[kubelet-start] Downloading configuration for the kubelet from the %q ConfigMap in the %s namespace",
		configMapName, metav1.NamespaceSystem)

	kubeletCfg, err := apiclient.GetConfigMapWithRetry(client, metav1.NamespaceSystem, configMapName)
	// If the ConfigMap wasn't found and the kubelet version is v1.10.x, where we didn't support the config file yet
	// just return, don't error out
	if apierrors.IsNotFound(err) && kubeletVersion.Minor() == 10 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// populate static pod path of kubelet configuration in Bhojpur DCP
	_, kubeletCodecs, err := kubeletscheme.NewSchemeAndCodecs()
	if err != nil {
		return nil, err
	}
	kc, err := kubeletcodec.DecodeKubeletConfiguration(kubeletCodecs, []byte(kubeletCfg.Data[constants.KubeletBaseConfigurationConfigMapKey]))
	if err != nil {
		return nil, err
	}
	if kc.StaticPodPath == "" {
		kc.StaticPodPath = filepath.Join(constants.KubernetesDir, constants.ManifestsSubDirName)
	}

	data, err := kubeletcodec.EncodeKubeletConfig(kc, kubeletconfigv1beta1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}

	return kc, writeConfigBytesToDisk(data, kubeletDir)
}

// writeConfigBytesToDisk writes a byte slice down to disk at the specific location of the kubelet config file
func writeConfigBytesToDisk(b []byte, kubeletDir string) error {
	configFile := filepath.Join(kubeletDir, constants.KubeletConfigurationFileName)
	klog.Infof("[kubelet-start] Writing kubelet configuration to file %q", configFile)

	// creates target folder if not already exists
	if err := os.MkdirAll(kubeletDir, 0700); err != nil {
		return errors.Wrapf(err, "failed to create directory %q", kubeletDir)
	}

	if err := ioutil.WriteFile(configFile, b, 0644); err != nil {
		return errors.Wrapf(err, "failed to write kubelet configuration to the file %q", configFile)
	}
	return nil
}

// addEngineStaticYaml generate Bhojpur DCP server engine static yaml for worker node.
func addEngineStaticYaml(data joindata.DcpJoinData, podManifestPath string) error {
	klog.Info("[join-node] Adding edge hub static yaml")
	if _, err := os.Stat(podManifestPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(podManifestPath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			klog.Errorf("Describe dir %s fail: %v", podManifestPath, err)
			return err
		}
	}

	ctx := map[string]string{
		"kubernetesServerAddr": fmt.Sprintf("https://%s", data.ServerAddr()),
		"image":                data.EngineImage(),
		"joinToken":            data.JoinToken(),
		"workingMode":          data.NodeRegistration().WorkingMode,
		"organizations":        data.NodeRegistration().Organizations,
	}

	engineTemplate, err := templates.SubsituteTemplate(edgenode.EngineTemplate, ctx)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(podManifestPath, dcpconstants.EngineStaticPodFileName), []byte(engineTemplate), 0600); err != nil {
		return err
	}
	klog.Info("[join-node] Add hub agent static yaml is ok")
	return nil
}
