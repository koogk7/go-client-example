package main

import (
	"encoding/json"
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	watch, err := clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	go func() {
		for event := range watch.ResultChan() {
			fmt.Printf("Type: %v\n", event.Type)
			p, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("unexpected type")
			}
			fmt.Println(p.Status.ContainerStatuses)
			fmt.Println(p.Status.Phase)
		}
		done <- true
	}()
	<-done
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	print(string(s))
	return string(s)
}
