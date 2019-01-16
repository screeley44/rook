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

package objectbucketclaim

// TODO clean out remaing ceph references

import (
	"fmt"
	"reflect"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/s3"

	"github.com/coreos/pkg/capnslog"
	opkit "github.com/rook/operator-kit"
	"k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"

	rookv1alpha2 "github.com/rook/rook/pkg/apis/rook.io/v1alpha2"
	"github.com/rook/rook/pkg/client/informers/externalversions"
	listersv1alpha2 "github.com/rook/rook/pkg/client/listers/rook.io/v1alpha2"
	"github.com/rook/rook/pkg/clusterd"
)

const (
	accessKey       string = "accessKeyId"
	secretKey       string = "secretAccessKey"
	crdNameSingular string = "objectbucketclaim"
	crdNamePlural   string = "objectbucketclaims"
)

// ObjectBucketResource represent the object store user custom resource for the watcher
var ObjectBucketResource = opkit.CustomResource{
	Name:    crdNameSingular,
	Plural:  crdNamePlural,
	Group:   rookv1alpha2.CustomResourceGroup,
	Version: rookv1alpha2.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(rookv1alpha2.ObjectBucketClaim{}).Name(),
}

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "object-bucket-claim")

// Controller encapsulates the object bucket controller event handling
type Controller struct {
	context      *clusterd.Context
	bucketLister listersv1alpha2.ObjectBucketClaimLister
	secretLister listers.SecretLister
}

// NewObjectBucketController create controller for watching object bucket custom resources
// A shared indexer is created to reduce calls to the API server.
func NewController(context *clusterd.Context) *Controller {
	externalInformerFactory := externalversions.NewSharedInformerFactory(context.RookClientset, 0)
	internalInformerFactory := informers.NewSharedInformerFactory(context.Clientset, 0)
	return &Controller{
		context:      context,
		bucketLister: externalInformerFactory.Rook().V1alpha2().ObjectBucketClaims().Lister(),
		secretLister: internalInformerFactory.Core().V1().Secrets().Lister(),
	}

}

// StartWatch watches CephObjectBucket custom resources and acts on API events
func (c *Controller) StartWatch(namespace string, stopCh <-chan struct{}) {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}
	logger.Infof("start watching object bucket resources")
	watcher := opkit.NewWatcher(ObjectBucketResource, namespace, resourceHandlerFuncs, c.context.RookClientset.RookV1alpha2().RESTClient())

	go watcher.Watch(&rookv1alpha2.ObjectBucketClaim{}, stopCh)
	return
}

func (c *Controller) onAdd(obj interface{}) {
	objectBucket, err := getObjectBucketResource(obj)
	if err != nil {
		logger.Errorf("failed to get CephObjectBucket resource: %v", err)
		return
	}
	c.handleAdd(objectBucket)
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	// TODO
}

func (c *Controller) onDelete(obj interface{}) {
	// TODO
}


// TODO scrap this
type cephUser struct {
	name, accessKey, secretKey string
}

func (c *Controller) handleAdd(obc *rookv1alpha2.ObjectBucketClaim) {
	var err error
	// get user and credentials
	user, err := c.newCephUserFromObjectBucket(obc)
	if err != nil {
		logger.Errorf("failed to retrieve ceph user: %v", err)
		return
	}
	var _ = user

	// s3-sdk create bucket with user/creds
	// createCephBucket(user, ob.Name)

	// create configMap
}

func (c *Controller) handleUpdate(ob *rookv1alpha2.ObjectBucketClaim) {
	// TODO
}

func (c *Controller) handleDelete(ob *rookv1alpha2.ObjectBucketClaim) {
	// TODO
}

func getObjectBucketResource(obj interface{}) (*rookv1alpha2.ObjectBucketClaim, error) {
	var ok bool
	objectBucket, ok := obj.(*rookv1alpha2.ObjectBucketClaim)
	if !ok {
		return nil, fmt.Errorf("not a known cephobjectbucket object: %+v", obj)
	}
	return objectBucket.DeepCopy(), nil
}

// TODO return access and secret key strings
// TODO parse secretName, get secret data
func (c *Controller) newCephUserFromObjectBucket(ob *rookv1alpha2.ObjectBucketClaim) (*cephUser, error) {
	var err error
	cu := &cephUser{}

	if cu.name, err = c.getObjectBucketUserFromBucket(ob); err != nil {
		return nil, fmt.Errorf("failed getting user name, %v", err)
	}
	if cu.accessKey, cu.secretKey, err = c.getObectBucketUserCredentials(ob); err != nil {
		return nil, fmt.Errorf("failed getting user keys: %v", err)
	}
	return cu, nil
}

func (c *Controller) getObjectBucketUserFromBucket(ob *rookv1alpha2.ObjectBucketClaim) (string, error) {
	user := ob.Spec.SecretName
	if user == "" {
		return "", fmt.Errorf("ObjectUser is empty, required")
	}
	return user, nil
}

func (c *Controller) getObectBucketUserCredentials(ob *rookv1alpha2.ObjectBucketClaim) (string, string, error) {
	secret, err := c.getObjectBucketUserSecret(ob.Name, ob.Namespace)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user credentials")
	}
	if len(secret.Data[accessKey]) == 0 {
		return "", "", fmt.Errorf("failed to get AccessKey, secret.data.AccessKey is empty")
	}
	if len(secret.Data[secretKey]) == 0 {
		return "", "", fmt.Errorf("failed to get SecretKey, secret.data.SecretKey is empty")
	}
	return fmt.Sprintf("%s", secret.Data[accessKey]), fmt.Sprintf("%s", secret.Data[secretKey]), nil
}

func (c *Controller) getObjectBucketUserSecret(user, namespace string) (*v1.Secret, error) {
	req, err := labels.NewRequirement("user", selection.Equals, []string{user})
	if err != nil {
		return nil, fmt.Errorf("error creating label requirement: %v", err)
	}
	secretList, err := c.secretLister.Secrets(namespace).List(labels.NewSelector().Add(*req))
	if err != nil {
		return nil, fmt.Errorf("could not list user secrets by 'user' label: %v", err)
	}
	if n := len(secretList); n != 1 {
		// unexpected edge case. A ceph object user can only have 1 set of access keys.  If < 1 <  secrets exist we
		// will not attempt to determine which one is legitimate.
		return nil, fmt.Errorf("expected to find 1 secret for namespace/user %s/%s, got %d", namespace, user, n)
	}
	return secretList[0], nil
}

// func createCephBucket(user *cephUser, bucketName string) error {
// 	s3Client := newS3ClientFromCreds(user.accessKey, user.secretKey)
// 	output, err := s3Client.CreateBucket(&s3.CreateBucketInput{
// 		Bucket: &bucketName,
// 	})
// 	logger.Infof("bucket creation output: %v", output)
// 	if err != nil {
// 		return fmt.Errorf("failed creating bucket %q", bucketName, err)
// 	}
// 	return nil
// }
//
// func newS3ClientFromCreds(id, secret string) *s3.S3 {
// 	sess := session.Must(session.NewSession(&aws.Config{
// 		Credentials: credentials.NewStaticCredentials(id, secret, ""),
// 	}))
// 	return s3.New(sess, aws.NewConfig())
// }
//
// // TODO
// func newCephBucketConfigMap(ob *cephv1beta1.CephObjectBucket) *v1.ConfigMap {
// 	return &v1.ConfigMap{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: rookCephPrefix + ob.Name,
// 		},
// 		Data: map[string]string{
// 			"BUCKET_HOST": "",
// 			"BUCKET_PORT": "",
// 			"BUCKET_NAME": "",
// 			"BUCKET_SSL":  "",
// 		},
// 	}
// }
