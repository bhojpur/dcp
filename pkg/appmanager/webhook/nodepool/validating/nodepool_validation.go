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
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
)

// annotationValidator validates the NodePool.Spec.Annotations
var annotationValidator = func(annos map[string]string) error {
	errs := apivalidation.ValidateAnnotations(annos, field.NewPath("field"))
	if len(errs) > 0 {
		return errors.New(errs.ToAggregate().Error())
	}
	return nil
}

func validateNodePoolSpecAnnotations(annotations map[string]string) field.ErrorList {
	if err := annotationValidator(annotations); err != nil {
		return field.ErrorList([]*field.Error{
			field.Invalid(field.NewPath("spec").Child("annotations"),
				annotations, "invalid annotations")})
	}
	return nil
}

// validateNodePoolSpec validates the nodepool spec.
func validateNodePoolSpec(spec *appsv1alpha1.NodePoolSpec) field.ErrorList {
	if allErrs := validateNodePoolSpecAnnotations(spec.Annotations); allErrs != nil {
		return allErrs
	}
	return nil
}

// validateNodePoolSpecUpdate tests if required fields in the NodePool spec are set.
func validateNodePoolSpecUpdate(spec, oldSpec *appsv1alpha1.NodePoolSpec) field.ErrorList {
	if allErrs := validateNodePoolSpec(spec); allErrs != nil {
		return allErrs
	}

	if spec.Type != oldSpec.Type {
		return field.ErrorList([]*field.Error{
			field.Invalid(field.NewPath("spec").Child("type"),
				spec.Annotations, "pool type can't be changed")})
	}
	return nil
}

// validateNodePoolDeletion validate the nodepool deletion event, which prevents
// the default-nodepool from being deleted
func validateNodePoolDeletion(cli client.Client, np *appsv1alpha1.NodePool) field.ErrorList {
	nodes := corev1.NodeList{}

	if np.Name == appsv1alpha1.DefaultCloudNodePoolName || np.Name == appsv1alpha1.DefaultEdgeNodePoolName {
		return field.ErrorList([]*field.Error{
			field.Forbidden(field.NewPath("metadata").Child("name"),
				fmt.Sprintf("default nodepool %s forbiden to delete", np.Name))})
	}

	if err := cli.List(context.TODO(), &nodes,
		client.MatchingLabels(np.Spec.Selector.MatchLabels)); err != nil {
		return field.ErrorList([]*field.Error{
			field.Forbidden(field.NewPath("metadata").Child("name"),
				"fail to get nodes associated to the pool")})
	}
	if len(nodes.Items) != 0 {
		return field.ErrorList([]*field.Error{
			field.Forbidden(field.NewPath("metadata").Child("name"),
				"cannot remove nonempty pool, please drain the pool before deleting")})
	}
	return nil
}
