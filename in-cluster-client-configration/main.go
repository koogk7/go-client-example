package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

func main() {
	config, err := rest.InClusterConfig()
	validate(err)

	clientset, err := kubernetes.NewForConfig(config)
	validate(err)

	for {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		validate(err)
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		_, err = clientset.CoreV1().Pods("default").Get("hello-node-rc", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found example-xxxxx pod in default namespace\n")
		}

		time.Sleep(10 * time.Second)
	}
}

func validate(err error) {
	if err != nil {
		panic(err)
	}
}
