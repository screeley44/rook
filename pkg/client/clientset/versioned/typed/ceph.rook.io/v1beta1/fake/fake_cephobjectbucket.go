/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCephObjectBuckets implements CephObjectBucketInterface
type FakeCephObjectBuckets struct {
	Fake *FakeCephV1beta1
	ns   string
}

var cephobjectbucketsResource = schema.GroupVersionResource{Group: "ceph.rook.io", Version: "v1beta1", Resource: "cephobjectbuckets"}

var cephobjectbucketsKind = schema.GroupVersionKind{Group: "ceph.rook.io", Version: "v1beta1", Kind: "CephObjectBucket"}

// Get takes name of the cephObjectBucket, and returns the corresponding cephObjectBucket object, and an error if there is any.
func (c *FakeCephObjectBuckets) Get(name string, options v1.GetOptions) (result *v1beta1.CephObjectBucket, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cephobjectbucketsResource, c.ns, name), &v1beta1.CephObjectBucket{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CephObjectBucket), err
}

// List takes label and field selectors, and returns the list of CephObjectBuckets that match those selectors.
func (c *FakeCephObjectBuckets) List(opts v1.ListOptions) (result *v1beta1.CephObjectBucketList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cephobjectbucketsResource, cephobjectbucketsKind, c.ns, opts), &v1beta1.CephObjectBucketList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.CephObjectBucketList{ListMeta: obj.(*v1beta1.CephObjectBucketList).ListMeta}
	for _, item := range obj.(*v1beta1.CephObjectBucketList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cephObjectBuckets.
func (c *FakeCephObjectBuckets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cephobjectbucketsResource, c.ns, opts))

}

// Create takes the representation of a cephObjectBucket and creates it.  Returns the server's representation of the cephObjectBucket, and an error, if there is any.
func (c *FakeCephObjectBuckets) Create(cephObjectBucket *v1beta1.CephObjectBucket) (result *v1beta1.CephObjectBucket, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cephobjectbucketsResource, c.ns, cephObjectBucket), &v1beta1.CephObjectBucket{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CephObjectBucket), err
}

// Update takes the representation of a cephObjectBucket and updates it. Returns the server's representation of the cephObjectBucket, and an error, if there is any.
func (c *FakeCephObjectBuckets) Update(cephObjectBucket *v1beta1.CephObjectBucket) (result *v1beta1.CephObjectBucket, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cephobjectbucketsResource, c.ns, cephObjectBucket), &v1beta1.CephObjectBucket{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CephObjectBucket), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCephObjectBuckets) UpdateStatus(cephObjectBucket *v1beta1.CephObjectBucket) (*v1beta1.CephObjectBucket, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(cephobjectbucketsResource, "status", c.ns, cephObjectBucket), &v1beta1.CephObjectBucket{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CephObjectBucket), err
}

// Delete takes name of the cephObjectBucket and deletes it. Returns an error if one occurs.
func (c *FakeCephObjectBuckets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cephobjectbucketsResource, c.ns, name), &v1beta1.CephObjectBucket{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCephObjectBuckets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cephobjectbucketsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.CephObjectBucketList{})
	return err
}

// Patch applies the patch and returns the patched cephObjectBucket.
func (c *FakeCephObjectBuckets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CephObjectBucket, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cephobjectbucketsResource, c.ns, name, data, subresources...), &v1beta1.CephObjectBucket{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CephObjectBucket), err
}
