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
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"

	rook "github.com/rook/rook/pkg/apis/rook.io/v1alpha2"
	"github.com/rook/rook/pkg/client/informers/externalversions"
	listersv1alpha2 "github.com/rook/rook/pkg/client/listers/rook.io/v1alpha2"
	"github.com/rook/rook/pkg/clusterd"
)

const (
	accessKeyId     string = "accessKeyId"
	secretAccessKey string = "secretAccessKey"
	crdNameSingular string = "objectbucketclaim"
	crdNamePlural   string = "objectbucketclaims"
)

// ObjectBucketResource represent the object store user custom resource for the watcher
var ObjectBucketResource = opkit.CustomResource{
	Name:    crdNameSingular,
	Plural:  crdNamePlural,
	Group:   rook.CustomResourceGroup,
	Version: rook.Version,
	Scope:   apiextensionsv1beta1.NamespaceScoped,
	Kind:    reflect.TypeOf(rook.ObjectBucketClaim{}).Name(),
}

var logger = capnslog.NewPackageLogger("github.com/rook/rook", "object-bucket-claim-operator")

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

// StartWatch watches ObjectBucketClaim custom resources and reacts to API events
func (c *Controller) StartWatch(namespace string, stopCh <-chan struct{}) {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}
	logger.Infof("start watching object bucket resources")
	watcher := opkit.NewWatcher(ObjectBucketResource, namespace, resourceHandlerFuncs, c.context.RookClientset.RookV1alpha2().RESTClient())
	go watcher.Watch(&rook.ObjectBucketClaim{}, stopCh)
	return
}

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

func (c *Controller) handleAdd(obc *rook.ObjectBucketClaim) {
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

func (c *Controller) handleUpdate(obc *rook.ObjectBucketClaim) {
	// TODO
}

func (c *Controller) handleDelete(obc *rook.ObjectBucketClaim) {
	// TODO
}

func copyObjectBucketClaim(obj interface{}) (*rook.ObjectBucketClaim, error) {
	var ok bool
	obc, ok := obj.(*rook.ObjectBucketClaim)
	if !ok {
		return nil, fmt.Errorf("type check failed, not an ObjectBucketClaim: %+v", obj)
	}
	return obc.DeepCopy(), nil
}

func (c *Controller) getAccessKeys(obc *rook.ObjectBucketClaim) (string, string, error) {
	secret, err := c.getAccessKeySecret(obc.Spec.SecretName, obc.Namespace)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user credentials")
	}
	ak := string(secret.Data[accessKeyId])
	sk := string(secret.Data[secretAccessKey])
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
