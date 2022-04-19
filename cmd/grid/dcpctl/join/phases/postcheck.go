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
	"net/http"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	"github.com/bhojpur/dcp/cmd/grid/dcpctl/join/joindata"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/options"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/cmd/phases/workflow"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/constants"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/apiclient"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/initsystem"
	"github.com/bhojpur/dcp/pkg/client/kubernetes/kubeadm/app/util/kubeconfig"
	"github.com/bhojpur/dcp/pkg/client/util/edgenode"
)

//NewPostcheckPhase creates a client workflow phase that check the health status of node components.
func NewPostcheckPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "postcheck",
		Short: "postcheck",
		Run:   runPostCheck,
		InheritFlags: []string{
			options.TokenStr,
		},
	}
}

// runPostCheck executes the node health check process.
func runPostCheck(c workflow.RunData) error {
	j, ok := c.(joindata.DcpJoinData)
	if !ok {
		return fmt.Errorf("Postcheck edge-node phase invoked with an invalid data struct. ")
	}

	klog.V(1).Infof("check kubelet status.")
	if err := checkKubeletStatus(); err != nil {
		return err
	}
	klog.V(1).Infof("kubelet service is active")

	klog.V(1).Infof("waiting hub agent ready.")
	if err := checkEngineHealthz(); err != nil {
		return err
	}
	klog.V(1).Infof("hub agent is ready")

	nodeRegistration := j.NodeRegistration()
	return patchNode(nodeRegistration.Name, nodeRegistration.CRISocket)
}

//checkKubeletStatus check if kubelet is healthy.
func checkKubeletStatus() error {
	initSystem, err := initsystem.GetInitSystem()
	if err != nil {
		return err
	}
	if ok := initSystem.ServiceIsActive("kubelet"); !ok {
		return fmt.Errorf("kubelet is not active. ")
	}
	return nil
}

//checkEngineHealthz check if Bhojpur DCP server engine is healthy.
func checkEngineHealthz() error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s%s", edgenode.ServerHealthzServer, edgenode.ServerHealthzURLPath), nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	return wait.PollImmediate(time.Second*5, 300*time.Second, func() (bool, error) {
		resp, err := client.Do(req)
		if err != nil {
			return false, nil
		}
		ok, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, nil
		}
		return string(ok) == "OK", nil
	})
}

//patchNode patch annotations for worker node.
func patchNode(nodeName, criSocket string) error {
	client, err := kubeconfig.ClientSetFromFile(filepath.Join(constants.KubernetesDir, constants.KubeletKubeConfigFileName))
	if err != nil {
		return err
	}

	return apiclient.PatchNode(client, nodeName, func(n *v1.Node) {
		if n.ObjectMeta.Annotations == nil {
			n.ObjectMeta.Annotations = make(map[string]string)
		}
		n.ObjectMeta.Annotations[constants.AnnotationKubeadmCRISocket] = criSocket
	})
}
