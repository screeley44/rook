/*
Copyright 2016 The Rook Authors. All rights reserved.

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

package objectbucket

import (
	"fmt"
	"reflect"

	"github.com/coreos/pkg/capnslog"
	opkit "github.com/rook/operator-kit"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cephv1beta1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1beta1"
	"github.com/rook/rook/pkg/client/informers/externalversions"
	obListerv1beta1 "github.com/rook/rook/pkg/client/listers/ceph.rook.io/v1beta1"
	"github.com/rook/rook/pkg/clusterd"
)

// ObjectBucketResource represent the object store user custom resource for the watcher
var ObjectBucketResource = opkit.CustomResource{
	Name:    "cephobjectbucket",
	Plural:  "cephobjectbucket",
	Group:   cephv1beta1.CustomResourceGroup,
	Version: cephv1beta1.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(cephv1beta1.ObjectBucket{}).Name(),
}

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "op-object")

// Controller encapsulates the object bucket controller event handling
type Controller struct {
	context  *clusterd.Context
	ownerRef metav1.OwnerReference
	lister   obListerv1beta1.ObjectBucketLister
}

// NewObjectBucketController create controller for watching object bucket custom resources
// A shared indexer is created to reduce calls to the API server.
func NewObjectBucketController(context *clusterd.Context, ownerRef metav1.OwnerReference) *Controller {
	informerFactory := externalversions.NewSharedInformerFactory(context.RookClientset, 0)
	return &Controller{
		context:  context,
		ownerRef: ownerRef,
		lister:   informerFactory.Ceph().V1beta1().ObjectBuckets().Lister(),
	}
}

// StartWatch watches ObjectBucket custom resources and acts on API events
func (c *Controller) StartWatch(stopCh <-chan struct{}) {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	logger.Infof("start watching object bucket resources")
	watcher := opkit.NewWatcher(ObjectBucketResource, "", resourceHandlerFuncs, c.context.RookClientset.CephV1beta1().RESTClient())

	go watcher.Watch(&cephv1beta1.ObjectBucket{}, stopCh)
	return
}

// onAdd handles new ObjectBucket events
func (c *Controller) onAdd(obj interface{}) {
	objectBucket, _, err := getObjectBucketObject(obj)
	if err != nil {
		logger.Errorf("failed to get objectbucket object: %v", err)
		return
	}
	// TODO
	var _ = objectBucket
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	// TODO
}

func (c *Controller) onDelete(obj interface{}) {
	// TODO
}

func getObjectBucketObject(obj interface{}) (*cephv1beta1.ObjectBucket, bool, error) {
	var ok bool
	objectBucket, ok := obj.(cephv1beta1.ObjectBucket)
	if ok {
		return objectBucket.DeepCopy(), false, nil
	}
	return nil, false, fmt.Errorf("obj does not match ObjectBucket type")
}

func (c *Controller) handleAdd(ob *cephv1beta1.ObjectBucket) {
	// TODO
}

func (c *Controller) handleUpdate(ob *cephv1beta1.ObjectBucket) {
	// TODO
}

func (c *Controller) handleDelete(ob *cephv1beta1.ObjectBucket) {
	// TODO
}
