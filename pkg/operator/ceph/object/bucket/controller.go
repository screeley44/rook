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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/coreos/pkg/capnslog"
	opkit "github.com/rook/operator-kit"
	"k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiserver/pkg/storage/names"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"

	// rook "github.com/rook/rook/pkg/apis/rook.io/v1alpha2" <-- unused in this branch
	cephv1beta1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1beta1"
	"github.com/rook/rook/pkg/client/informers/externalversions"
	obListerv1beta1 "github.com/rook/rook/pkg/client/listers/ceph.rook.io/v1beta1"
	"github.com/rook/rook/pkg/clusterd"
)

const (
	rookCephPrefix  string = "rook-ceph-object-bucket-"
	accessKey       string = "AccessKey"
	secretKey       string = "SecretKey"
	crdNameSingular string = "cephobjectbucket"
	crdNamePlural   string = "cephobjectbuckets"
	//accessKeyId     string = "accessKeyId"
	//secretAccessKey string = "secretAccessKey"
	//crdNameSingular string = "objectbucketclaim"
	//crdNamePlural   string = "objectbucketclaims"
)

// ObjectBucketResource represent the object store user custom resource for the watcher
var ObjectBucketResource = opkit.CustomResource{
	Name:    crdNameSingular,
	Plural:  crdNamePlural,
	Group:   cephv1beta1.CustomResourceGroup,
	Version: cephv1beta1.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(cephv1beta1.CephObjectBucket{}).Name(),
}

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "op-object")

// Controller encapsulates the object bucket controller event handling
type Controller struct {
	context      *clusterd.Context
	ownerRef     metav1.OwnerReference
	bucketLister obListerv1beta1.CephObjectBucketLister
	secretLister listers.SecretLister
}

// NewObjectBucketController create controller for watching object bucket custom resources
// A shared indexer is created to reduce calls to the API server.
func NewObjectBucketController(context *clusterd.Context, ownerRef metav1.OwnerReference) *Controller {
	externalInformerFactory := externalversions.NewSharedInformerFactory(context.RookClientset, 0)
	internalInformerFactory := informers.NewSharedInformerFactory(context.Clientset, 0)
	return &Controller{
		context:      context,
		ownerRef:     ownerRef,
		bucketLister: externalInformerFactory.Ceph().V1beta1().CephObjectBuckets().Lister(),
		secretLister: internalInformerFactory.Core().V1().Secrets().Lister(),
	}

}

// StartWatch watches CephObjectBucket custom resources and acts on API events
func (c *Controller) StartWatch(stopCh <-chan struct{}) {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}
	logger.Infof("start watching object bucket resources")
	watcher := opkit.NewWatcher(ObjectBucketResource, "", resourceHandlerFuncs, c.context.RookClientset.CephV1beta1().RESTClient())

	go watcher.Watch(&cephv1beta1.CephObjectBucket{}, stopCh)
	return
}

// Replaced by obc-controller branch - keeping for reference
//func (c *Controller) onAdd(obj interface{}) {
//	objectBucket, err := getObjectBucketResource(obj)
//	if err != nil {
//		logger.Errorf("failed to get CephObjectBucket resource: %v", err)
//		return
//	}
//	c.handleAdd(objectBucket)
//}

// From obc-controller branch
func (c *Controller) onAdd(obj interface{}) {
	obc, err := copyObjectBucketClaim(obj)
	if err != nil {
		logger.Errorf("failed to get ObjectBucketClaim resource: %v", err)
		return
	}
	c.handleAdd(obc)
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	// TODO
}

func (c *Controller) onDelete(obj interface{}) {
	// TODO
}

type cephUser struct {
	name, accessKey, secretKey string
}

// From obc-controller branch
func (c *Controller) handleAdd(obc *cephv1beta1.CephObjectBucket) {
	accessKey, secretKey, err := c.getAccessKeys(obc)
	if err != nil {
		logger.Errorf("failed to extract access keys: %v", err)
		return
	}

	bucketName := names.SimpleNameGenerator.GenerateName(obc.Spec.GenerateBucketName)

	err = createBucket(bucketName, accessKey, secretKey)
	if err != nil {
		logger.Errorf("could not create bucket: %v", err)
	}
	// create configMap
}

// Replaced by obc-controller branch - keeping for reference
//func (c *Controller) handleAdd(ob *cephv1beta1.CephObjectBucket) {
//	var err error
//	// get user and credentials
//	user, err := c.newCephUserFromObjectBucket(ob)
//	if err != nil {
//		logger.Errorf("failed to retrieve ceph user: %v", err)
//		return
//	}
//	var _ = user
//
//	// s3-sdk create bucket with user/creds
//	// createCephBucket(user, ob.Name)
//
//	// create configMap
//}

func (c *Controller) handleUpdate(ob *cephv1beta1.CephObjectBucket) {
	// TODO
}

func (c *Controller) handleDelete(ob *cephv1beta1.CephObjectBucket) {
	// TODO
}

func copyObjectBucketClaim(obj interface{}) (*cephv1beta1.CephObjectBucket, error) {
	var ok bool
	obc, ok := obj.(*cephv1beta1.CephObjectBucket)
	if !ok {
		return nil, fmt.Errorf("type check failed, not an ObjectBucketClaim: %+v", obj)
	}
	return obc.DeepCopy(), nil
}

//TODO maybe can be removed since it's not used - but keeping as reference for now?
func getObjectBucketResource(obj interface{}) (*cephv1beta1.CephObjectBucket, error) {
	var ok bool
	objectBucket, ok := obj.(*cephv1beta1.CephObjectBucket)
	if !ok {
		return nil, fmt.Errorf("not a known cephobjectbucket object: %+v", obj)
	}
	return objectBucket.DeepCopy(), nil
}

func (c *Controller) newCephUserFromObjectBucket(ob *cephv1beta1.CephObjectBucket) (*cephUser, error) {
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

func (c *Controller) getObjectBucketUserFromBucket(ob *cephv1beta1.CephObjectBucket) (string, error) {
	user := ob.Spec.ObjectUser
	if user == "" {
		return "", fmt.Errorf("ObjectUser is empty, required")
	}
	return user, nil
}

func (c *Controller) getObectBucketUserCredentials(ob *cephv1beta1.CephObjectBucket) (string, string, error) {
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

func (c *Controller) getAccessKeys(obc *cephv1beta1.CephObjectBucket) (string, string, error) {
	//TODO - obc.Spec.BucketName isn't right - need to add status.secretName field or something to CephObjectBucket
	secret, err := c.getAccessKeySecret(obc.Spec.BucketName, obc.Namespace)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user credentials")
	}
	ak := string(secret.Data[accessKey])
	sk := string(secret.Data[secretKey])
	if len(ak) == 0 || len(sk) == 0 {
		return "", "", fmt.Errorf("invalid secret key value, key(s) cannot be empty")
	}
	return ak, sk, nil
}

func (c *Controller) getAccessKeySecret(secretName, namespace string) (*v1.Secret, error) {
	sec, err := c.secretLister.Secrets(namespace).Get(secretName)
	if err != nil {
		return nil, fmt.Errorf("could not list user secrets by 'user' label: %v", err)
	}
	return sec, nil
}

func createBucket(bucketName, accessKey, secretKey string) error {
	// Instantiate a short lived S3 client
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))
	s3c := s3.New(sess)

	output, err := s3c.CreateBucket(&s3.CreateBucketInput{
		Bucket: &bucketName,
	})
	logger.Infof("bucket creation output: %v", output)
	if err != nil {
		return fmt.Errorf("failed creating bucket %q: %v", bucketName, err)
	}
	return nil
}

// From obc-controller branch
//
// // TODO
// func newBucketConfigMap() *v1.ConfigMap {
// 	return &v1.ConfigMap{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: rookPrefix + ob.Name,
// 		},
// 		Data: map[string]string{
// 			"BUCKET_HOST": "",
// 			"BUCKET_PORT": "",
// 			"BUCKET_NAME": "",
// 			"BUCKET_SSL":  "",
// 		},
// 	}
// }

// From original object-bucket-controller branch
//TODO
// These commented out prior to branch merge - keeping as reference for now

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
