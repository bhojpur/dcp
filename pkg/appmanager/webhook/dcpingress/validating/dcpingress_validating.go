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

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

// validateIngressSpec validates the Bhojpur DCP ingress spec.
func validateIngressSpec(c client.Client, spec *appsv1alpha1.DcpIngressSpec) field.ErrorList {
	if len(spec.Pools) > 0 {
		var err error
		var errmsg string
		nps := appsv1alpha1.NodePoolList{}
		if err = c.List(context.TODO(), &nps, &client.ListOptions{}); err != nil {
			errmsg = "List nodepool list error!"
			klog.Errorf(errmsg)
			return field.ErrorList([]*field.Error{
				field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
		}

		// validate whether the nodepool exist
		if len(nps.Items) <= 0 {
			errmsg = "No nodepool is created in the cluster!"
			klog.Errorf(errmsg)
			return field.ErrorList([]*field.Error{
				field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
		} else {
			var found = false
			for _, snp := range spec.Pools { //go through the nodepools setting in yaml
				for _, cnp := range nps.Items { //go through the nodepools in cluster
					if snp.Name == cnp.ObjectMeta.Name {
						found = true
						break
					}
				}
				if !found {
					errmsg = snp.Name + " does not exist in the cluster!"
					klog.Errorf(errmsg)
					return field.ErrorList([]*field.Error{
						field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
				}
				found = false
			}

		}
	}
	return nil
}

func validateIngressSpecUpdate(c client.Client, spec, oldSpec *appsv1alpha1.DcpIngressSpec) field.ErrorList {
	return validateIngressSpec(c, spec)
}

func validateIngressSpecDeletion(c client.Client, spec *appsv1alpha1.DcpIngressSpec) field.ErrorList {
	return validateIngressSpec(c, spec)
}
