package mutating

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
	"encoding/json"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	"github.com/bhojpur/dcp/pkg/appmanager/util"
	webhookutil "github.com/bhojpur/dcp/pkg/appmanager/webhook/util"
)

// NodePoolCreateUpdateHandler handles UnitedDeployment
type NodePoolCreateUpdateHandler struct {
	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ webhookutil.Handler = &NodePoolCreateUpdateHandler{}

func (h *NodePoolCreateUpdateHandler) SetOptions(options webhookutil.Options) {
	return
}

// Handle handles admission requests.
func (h *NodePoolCreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	np := appsv1alpha1.NodePool{}
	err := h.Decoder.Decode(req, &np)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	klog.V(5).Infof("set the nodepool(%s) selector", np.Name)
	np.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{appsv1alpha1.LabelCurrentNodePool: np.Name},
	}

	marshalled, err := json.Marshal(&np)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	resp := admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw,
		marshalled)
	if len(resp.Patches) > 0 {
		klog.V(5).Infof("Admit NodePool %s patches: %v", np.Name, util.DumpJSON(resp.Patches))
	}
	return resp
}

var _ admission.DecoderInjector = &NodePoolCreateUpdateHandler{}

// InjectDecoder injects the decoder into the UnitedDeploymentCreateUpdateHandler
func (h *NodePoolCreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}
