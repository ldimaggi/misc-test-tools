/*
Copyright 2016 The Kubernetes Authors.

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

// Note: the example only works with the code within the same release/branch.

// Simple program to illustrate creating/deleting a namespace through client-go
// Original form of this program:
// https://github.com/kubernetes/client-go/tree/master/examples/out-of-cluster-client-configuration

// References:
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	CoreV1 "k8s.io/api/core/v1"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {

	var name string
	var labels map[string]string
	var isolation string
	var seconds *int64

	// type needed in creating a new namespace
	type namespacePolicy struct {
		Ingress struct {
			Isolation string `json:"isolation"`
		} `json:"ingress"`
	}

	// access kubeconfig
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// print out all existing namespaces
	nsList, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("There are %d namespaces in the cluster", len(nsList.Items))

	for _, ns := range nsList.Items {
		fmt.Printf("Namespace = %s\n", ns.ObjectMeta.Name)
	}

	// create a new namespace
	name = "my-test-namespace"
	ns_in := &CoreV1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
	np := namespacePolicy{}
	np.Ingress.Isolation = isolation
	annotation, _ := json.Marshal(np)
	ns_in.ObjectMeta.Annotations = map[string]string{
		"net.beta.kubernetes.io/network-policy": string(annotation),
	}

	ns, err := clientset.CoreV1().Namespaces().Create(ns_in)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Namespace = %s\n", ns.ObjectMeta.Name)

	_, err = clientset.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("namespace %s not found\n", name)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting namespace %s: %v\n",
			name, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found namespace %s\n", name)
	}

	// delete the recently created namespace
	deleteoptions := metav1.DeleteOptions{
		GracePeriodSeconds: seconds}

	err = clientset.CoreV1().Namespaces().Delete(ns.ObjectMeta.Name, &deleteoptions)
	if err != nil {
		panic(err)
	}

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
