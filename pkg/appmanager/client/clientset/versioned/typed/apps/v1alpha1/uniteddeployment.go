// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

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
	"time"

	v1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	scheme "github.com/bhojpur/dcp/pkg/appmanager/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// UnitedDeploymentsGetter has a method to return a UnitedDeploymentInterface.
// A group's client should implement this interface.
type UnitedDeploymentsGetter interface {
	UnitedDeployments(namespace string) UnitedDeploymentInterface
}

// UnitedDeploymentInterface has methods to work with UnitedDeployment resources.
type UnitedDeploymentInterface interface {
	Create(ctx context.Context, unitedDeployment *v1alpha1.UnitedDeployment, opts v1.CreateOptions) (*v1alpha1.UnitedDeployment, error)
	Update(ctx context.Context, unitedDeployment *v1alpha1.UnitedDeployment, opts v1.UpdateOptions) (*v1alpha1.UnitedDeployment, error)
	UpdateStatus(ctx context.Context, unitedDeployment *v1alpha1.UnitedDeployment, opts v1.UpdateOptions) (*v1alpha1.UnitedDeployment, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.UnitedDeployment, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.UnitedDeploymentList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.UnitedDeployment, err error)
	UnitedDeploymentExpansion
}

// unitedDeployments implements UnitedDeploymentInterface
type unitedDeployments struct {
	client rest.Interface
	ns     string
}

// newUnitedDeployments returns a UnitedDeployments
func newUnitedDeployments(c *AppsV1alpha1Client, namespace string) *unitedDeployments {
	return &unitedDeployments{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the unitedDeployment, and returns the corresponding unitedDeployment object, and an error if there is any.
func (c *unitedDeployments) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.UnitedDeployment, err error) {
	result = &v1alpha1.UnitedDeployment{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("uniteddeployments").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of UnitedDeployments that match those selectors.
func (c *unitedDeployments) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.UnitedDeploymentList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.UnitedDeploymentList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("uniteddeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested unitedDeployments.
func (c *unitedDeployments) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("uniteddeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a unitedDeployment and creates it.  Returns the server's representation of the unitedDeployment, and an error, if there is any.
func (c *unitedDeployments) Create(ctx context.Context, unitedDeployment *v1alpha1.UnitedDeployment, opts v1.CreateOptions) (result *v1alpha1.UnitedDeployment, err error) {
	result = &v1alpha1.UnitedDeployment{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("uniteddeployments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(unitedDeployment).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a unitedDeployment and updates it. Returns the server's representation of the unitedDeployment, and an error, if there is any.
func (c *unitedDeployments) Update(ctx context.Context, unitedDeployment *v1alpha1.UnitedDeployment, opts v1.UpdateOptions) (result *v1alpha1.UnitedDeployment, err error) {
	result = &v1alpha1.UnitedDeployment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("uniteddeployments").
		Name(unitedDeployment.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(unitedDeployment).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *unitedDeployments) UpdateStatus(ctx context.Context, unitedDeployment *v1alpha1.UnitedDeployment, opts v1.UpdateOptions) (result *v1alpha1.UnitedDeployment, err error) {
	result = &v1alpha1.UnitedDeployment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("uniteddeployments").
		Name(unitedDeployment.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(unitedDeployment).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the unitedDeployment and deletes it. Returns an error if one occurs.
func (c *unitedDeployments) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("uniteddeployments").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *unitedDeployments) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("uniteddeployments").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched unitedDeployment.
func (c *unitedDeployments) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.UnitedDeployment, err error) {
	result = &v1alpha1.UnitedDeployment{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("uniteddeployments").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}