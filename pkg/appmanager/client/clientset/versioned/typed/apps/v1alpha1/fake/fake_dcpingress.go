// Code generated by client-gen. DO NOT EDIT.

package fake

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

	v1alpha1 "github.com/bhojpur/dcp/pkg/appmanager/apis/apps/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeIngresses implements IngressInterface
type FakeIngresses struct {
	Fake *FakeAppsV1alpha1
}

var dcpingressesResource = schema.GroupVersionResource{Group: "apps.bhojpur.net", Version: "v1alpha1", Resource: "dcpingresses"}

var dcpingressesKind = schema.GroupVersionKind{Group: "apps.bhojpur.net", Version: "v1alpha1", Kind: "DcpIngress"}

// Get takes name of the dcpIngress, and returns the corresponding dcpIngress object, and an error if there is any.
func (c *FakeIngresses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.DcpIngress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(dcpingressesResource, name), &v1alpha1.DcpIngress{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DcpIngress), err
}

// List takes label and field selectors, and returns the list of DcpIngresses that match those selectors.
func (c *FakeIngresses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DcpIngressList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(dcpingressesResource, dcpingressesKind, opts), &v1alpha1.DcpIngressList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DcpIngressList{ListMeta: obj.(*v1alpha1.DcpIngressList).ListMeta}
	for _, item := range obj.(*v1alpha1.DcpIngressList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested dcpIngresses.
func (c *FakeIngresses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(dcpingressesResource, opts))
}

// Create takes the representation of a dcpIngress and creates it.  Returns the server's representation of the dcpIngress, and an error, if there is any.
func (c *FakeIngresses) Create(ctx context.Context, dcpIngress *v1alpha1.DcpIngress, opts v1.CreateOptions) (result *v1alpha1.DcpIngress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(dcpingressesResource, dcpIngress), &v1alpha1.DcpIngress{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DcpIngress), err
}

// Update takes the representation of a dcpIngress and updates it. Returns the server's representation of the dcpIngress, and an error, if there is any.
func (c *FakeIngresses) Update(ctx context.Context, dcpIngress *v1alpha1.DcpIngress, opts v1.UpdateOptions) (result *v1alpha1.DcpIngress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(dcpingressesResource, dcpIngress), &v1alpha1.DcpIngress{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DcpIngress), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeIngresses) UpdateStatus(ctx context.Context, dcpIngress *v1alpha1.DcpIngress, opts v1.UpdateOptions) (*v1alpha1.DcpIngress, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(dcpingressesResource, "status", dcpIngress), &v1alpha1.DcpIngress{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DcpIngress), err
}

// Delete takes name of the dcpIngress and deletes it. Returns an error if one occurs.
func (c *FakeIngresses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(dcpingressesResource, name), &v1alpha1.DcpIngress{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIngresses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(dcpingressesResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DcpIngressList{})
	return err
}

// Patch applies the patch and returns the patched dcpIngress.
func (c *FakeIngresses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.DcpIngress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(dcpingressesResource, name, pt, data, subresources...), &v1alpha1.DcpIngress{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.DcpIngress), err
}