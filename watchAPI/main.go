package main

import (
	"encoding/json"
	"flag"
	"fmt"
	apps_v1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
)

func main() {
	done := make(chan bool)
	//============= 쿠버네티스 config 파싱 ================
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) kubeconfig absolute path")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig path")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// kubernetes.NewForConfig 는 config에 맞는 클라이언트 셋을 리턴한다.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	label := "app=products"
	namespace := "default"

	watch, err := clientset.AppsV1beta1().Deployments(namespace).Watch(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	list, err := clientset.AppsV1beta1().Deployments(namespace).List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	resultList(*list)
	go resultWatch(watch, done)
	<-done
}

func resultWatch(watch watch.Interface, done chan<- bool) {
	log.Println("result watch")
	for event := range watch.ResultChan() {
		fmt.Printf("Type: %v\n", event.Type)
		eventObject, ok := event.Object.(*apps_v1beta1.Deployment)
		if !ok {
			log.Fatal("unexpected type")
		}
		prettyPrint(eventObject.Status)
	}
	done <- true
}

func resultList(list apps_v1beta1.DeploymentList) {
	log.Println("result list")
	for _, item := range list.Items {
		prettyPrint(item.Status)
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	println(string(s))
	return string(s)
}
