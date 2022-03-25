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

// IngressCreateUpdateHandler handles DcpIngress
type IngressCreateUpdateHandler struct {
	Client client.Client

	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ webhookutil.Handler = &IngressCreateUpdateHandler{}

func (h *IngressCreateUpdateHandler) SetOptions(options webhookutil.Options) {
	return
}

// Handle handles admission requests.
func (h *IngressCreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := appsv1alpha1.DcpIngress{}

	// singleton node pool validating
	if req.Name != appsv1alpha1.SingletonDcpIngressInstanceName {
		var msg = "please name DcpIngress with " + appsv1alpha1.SingletonDcpIngressInstanceName + " instead of " + req.Name
		klog.Errorf(msg)
		return admission.ValidationResponse(false, msg)
	}

	switch req.AdmissionRequest.Operation {
	case admissionv1.Create:
		klog.V(4).Info("capture the ingress creation request")

		if err := h.Decoder.Decode(req, &ingress); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if allErrs := validateIngressSpec(h.Client, &ingress.Spec); len(allErrs) > 0 {
			return admission.Errored(http.StatusUnprocessableEntity,
				allErrs.ToAggregate())
		}
	case admissionv1.Update:
		klog.V(4).Info("capture the ingress update request")
		if err := h.Decoder.Decode(req, &ingress); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oingress := appsv1alpha1.DcpIngress{}
		if err := h.Decoder.DecodeRaw(req.OldObject, &oingress); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if allErrs := validateIngressSpecUpdate(h.Client, &ingress.Spec, &oingress.Spec); len(allErrs) > 0 {
			return admission.Errored(http.StatusUnprocessableEntity,
				allErrs.ToAggregate())
		}
	case admissionv1.Delete:
		klog.V(4).Info("capture the ingress deletion request")
		if err := h.Decoder.DecodeRaw(req.OldObject, &ingress); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if allErrs := validateIngressSpecDeletion(h.Client, &ingress.Spec); len(allErrs) > 0 {
			return admission.Errored(http.StatusUnprocessableEntity,
				allErrs.ToAggregate())
		}
	}

	return admission.ValidationResponse(true, "")
}

var _ admission.DecoderInjector = &IngressCreateUpdateHandler{}

// InjectDecoder injects the decoder into the IngressCreateUpdateHandler
func (h *IngressCreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}

var _ inject.Client = &IngressCreateUpdateHandler{}

// InjectClient injects the client into the PodCreateHandler
func (h *IngressCreateUpdateHandler) InjectClient(c client.Client) error {
	h.Client = c
	return nil
}
