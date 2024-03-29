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

// NodePoolsGetter has a method to return a NodePoolInterface.
// A group's client should implement this interface.
type NodePoolsGetter interface {
	NodePools() NodePoolInterface
}

// NodePoolInterface has methods to work with NodePool resources.
type NodePoolInterface interface {
	Create(ctx context.Context, nodePool *v1alpha1.NodePool, opts v1.CreateOptions) (*v1alpha1.NodePool, error)
	Update(ctx context.Context, nodePool *v1alpha1.NodePool, opts v1.UpdateOptions) (*v1alpha1.NodePool, error)
	UpdateStatus(ctx context.Context, nodePool *v1alpha1.NodePool, opts v1.UpdateOptions) (*v1alpha1.NodePool, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.NodePool, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.NodePoolList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.NodePool, err error)
	NodePoolExpansion
}

// nodePools implements NodePoolInterface
type nodePools struct {
	client rest.Interface
}

// newNodePools returns a NodePools
func newNodePools(c *AppsV1alpha1Client) *nodePools {
	return &nodePools{
		client: c.RESTClient(),
	}
}

// Get takes name of the nodePool, and returns the corresponding nodePool object, and an error if there is any.
func (c *nodePools) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.NodePool, err error) {
	result = &v1alpha1.NodePool{}
	err = c.client.Get().
		Resource("nodepools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NodePools that match those selectors.
func (c *nodePools) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.NodePoolList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.NodePoolList{}
	err = c.client.Get().
		Resource("nodepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nodePools.
func (c *nodePools) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("nodepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a nodePool and creates it.  Returns the server's representation of the nodePool, and an error, if there is any.
func (c *nodePools) Create(ctx context.Context, nodePool *v1alpha1.NodePool, opts v1.CreateOptions) (result *v1alpha1.NodePool, err error) {
	result = &v1alpha1.NodePool{}
	err = c.client.Post().
		Resource("nodepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodePool).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a nodePool and updates it. Returns the server's representation of the nodePool, and an error, if there is any.
func (c *nodePools) Update(ctx context.Context, nodePool *v1alpha1.NodePool, opts v1.UpdateOptions) (result *v1alpha1.NodePool, err error) {
	result = &v1alpha1.NodePool{}
	err = c.client.Put().
		Resource("nodepools").
		Name(nodePool.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodePool).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *nodePools) UpdateStatus(ctx context.Context, nodePool *v1alpha1.NodePool, opts v1.UpdateOptions) (result *v1alpha1.NodePool, err error) {
	result = &v1alpha1.NodePool{}
	err = c.client.Put().
		Resource("nodepools").
		Name(nodePool.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodePool).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the nodePool and deletes it. Returns an error if one occurs.
func (c *nodePools) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("nodepools").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nodePools) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("nodepools").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched nodePool.
func (c *nodePools) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.NodePool, err error) {
	result = &v1alpha1.NodePool{}
	err = c.client.Patch(pt).
		Resource("nodepools").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}