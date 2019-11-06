/*
Copyright 2019 The Kubernetes my-crd-controller Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/nevermosby/my-crd-controller/pkg/apis/mycontroller/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WebsiteLister helps list Websites.
type WebsiteLister interface {
	// List lists all Websites in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Website, err error)
	// Websites returns an object that can list and get Websites.
	Websites(namespace string) WebsiteNamespaceLister
	WebsiteListerExpansion
}

// websiteLister implements the WebsiteLister interface.
type websiteLister struct {
	indexer cache.Indexer
}

// NewWebsiteLister returns a new WebsiteLister.
func NewWebsiteLister(indexer cache.Indexer) WebsiteLister {
	return &websiteLister{indexer: indexer}
}

// List lists all Websites in the indexer.
func (s *websiteLister) List(selector labels.Selector) (ret []*v1alpha1.Website, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Website))
	})
	return ret, err
}

// Websites returns an object that can list and get Websites.
func (s *websiteLister) Websites(namespace string) WebsiteNamespaceLister {
	return websiteNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WebsiteNamespaceLister helps list and get Websites.
type WebsiteNamespaceLister interface {
	// List lists all Websites in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Website, err error)
	// Get retrieves the Website from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Website, error)
	WebsiteNamespaceListerExpansion
}

// websiteNamespaceLister implements the WebsiteNamespaceLister
// interface.
type websiteNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Websites in the indexer for a given namespace.
func (s websiteNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Website, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Website))
	})
	return ret, err
}

// Get retrieves the Website from the indexer for a given namespace and name.
func (s websiteNamespaceLister) Get(name string) (*v1alpha1.Website, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("website"), name)
	}
	return obj.(*v1alpha1.Website), nil
}