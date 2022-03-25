package validating

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
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	webhookutil "github.com/bhojpur/dcp/pkg/appmanager/webhook/util"
)

// NodePoolCreateUpdateHandler handles UnitedDeployment
type NodePoolCreateUpdateHandler struct {
	Client client.Client

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

	switch req.AdmissionRequest.Operation {
	case admissionv1.Create:
		klog.V(4).Info("capture the nodepool creation request")
		err := h.Decoder.Decode(req, &np)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if allErrs := validateNodePoolSpec(&np.Spec); len(allErrs) > 0 {
			return admission.Errored(http.StatusUnprocessableEntity,
				allErrs.ToAggregate())
		}
	case admissionv1.Update:
		klog.V(4).Info("capture the nodepool update request")
		err := h.Decoder.Decode(req, &np)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		onp := appsv1alpha1.NodePool{}
		err = h.Decoder.DecodeRaw(req.OldObject, &onp)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if allErrs := validateNodePoolSpecUpdate(&np.Spec, &onp.Spec); len(allErrs) > 0 {
			return admission.Errored(http.StatusUnprocessableEntity,
				allErrs.ToAggregate())
		}
	case admissionv1.Delete:
		klog.V(4).Info("capture the nodepool deletion request")
		err := h.Decoder.DecodeRaw(req.OldObject, &np)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if allErrs := validateNodePoolDeletion(h.Client, &np); len(allErrs) > 0 {
			return admission.Errored(http.StatusUnprocessableEntity,
				allErrs.ToAggregate())
		}
	}

	return admission.ValidationResponse(true, "")
}

var _ admission.DecoderInjector = &NodePoolCreateUpdateHandler{}

// InjectDecoder injects the decoder into the UnitedDeploymentCreateUpdateHandler
func (h *NodePoolCreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}

var _ inject.Client = &NodePoolCreateUpdateHandler{}

// InjectClient injects the client into the PodCreateHandler
func (h *NodePoolCreateUpdateHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}
