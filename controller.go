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

// Package main for a sample operator
package main

import (
	"fmt"

	opkit "github.com/rook/operator-kit"
	sample "github.com/rook/operator-kit/sample-operator/pkg/apis/myproject/v1alpha1"
	sampleclient "github.com/jakewins/crd-missing-type/pkg/client/clientset/versioned/typed/myproject/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

// SampleController represents a controller object for sample custom resources
type SampleController struct {
	context         *opkit.Context
	sampleClientset sampleclient.MyprojectV1alpha1Interface
}

// newSampleController create controller for watching sample custom resources created
func newSampleController(context *opkit.Context, sampleClientset sampleclient.MyprojectV1alpha1Interface) *SampleController {
	return &SampleController{
		context:         context,
		sampleClientset: sampleClientset,
	}
}

// Watch watches for instances of Sample custom resources and acts on them
func (c *SampleController) StartWatch(namespace string, stopCh chan struct{}) error {

	resourceHandlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
	}
	restClient := c.sampleClientset.RESTClient()
	watcher := opkit.NewWatcher(sample.SampleResource, namespace, resourceHandlers, restClient)
	go watcher.Watch(&sample.Sample{}, stopCh)
	return nil
}

func (c *SampleController) onAdd(obj interface{}) {
	s := obj.(*sample.Sample).DeepCopy()

	if s.TypeMeta.Kind == "" {
		fmt.Printf("Oh no! Missing TypeMeta: found: %v", s)
		return
	}
	fmt.Printf("All good, Kind is %s.\n", s.TypeMeta.Kind)
}