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
	"os"
	"os/signal"
	"syscall"
	"time"

	opkit "github.com/rook/operator-kit"
	sample "github.com/jakewins/crd-missing-type/pkg/apis/myproject/v1alpha1"
	sampleapi "github.com/jakewins/crd-missing-type/pkg/client/clientset/versioned/typed/myproject/v1alpha1"
	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
	"flag"
)

func main() {
	fmt.Println("Getting kubernetes context")
	context, sampleClientset, err := createContext()
	if err != nil {
		fmt.Printf("failed to create context. %+v\n", err)
		os.Exit(1)
	}

	// Create and wait for CRD resources
	fmt.Println("Registering the sample resource")
	resources := []opkit.CustomResource{sample.SampleResource}
	err = opkit.CreateCustomResources(*context, resources)
	if err != nil {
		fmt.Printf("failed to create custom resource. %+v\n", err)
		os.Exit(1)
	}

	// create signals to stop watching the resources
	signalChan := make(chan os.Signal, 1)
	stopChan := make(chan struct{})
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// start watching the sample resource
	fmt.Println("Watching the sample resource")
	controller := newSampleController(context, sampleClientset)
	controller.StartWatch(v1.NamespaceAll, stopChan)

	for {
		select {
		case <-signalChan:
			fmt.Println("shutdown signal received, exiting...")
			close(stopChan)
			return
		}
	}
}

func createContext() (*opkit.Context, sampleapi.MyprojectV1alpha1Interface, error) {
	var cfg *rest.Config
	var err error
	if inClusterConfigAvailable() {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = outsideClusterConfig()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get k8s config. %+v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get k8s client. %+v", err)
	}

	apiExtClientset, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create k8s API extension clientset. %+v", err)
	}

	neo4jClientSet, err := sampleapi.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sample clientset. %+v", err)
	}

	context := &opkit.Context{
		Clientset:             clientset,
		APIExtensionClientset: apiExtClientset,
		Interval:              500 * time.Millisecond,
		Timeout:               60 * time.Second,
	}
	return context, neo4jClientSet, nil

}

func inClusterConfigAvailable() bool {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	return len(host) > 0 && len(port) > 0
}

func outsideClusterConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
