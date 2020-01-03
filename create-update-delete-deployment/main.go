package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	"os"
	"path/filepath"
)

/*Todo[todo]
개선사항
1. delete가 완전하게 안됨
2. pod list 불러오기 (지금은 deployemnt 만 불러옴)
3. 리팩토링
*/

func main() {
	metaName := "demo-deployment"
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
	//=====================================================

	//=============Deployment Controller 정의 ================
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault) // 파라미터로 네임스페이스를 받음

	deployment := &appsv1.Deployment{ // yaml 정의
		ObjectMeta: metav1.ObjectMeta{
			Name: metaName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "demo"},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "web",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	//=====================================================
	//=============Deployment Controller  생성 ================
	fmt.Println("Creating deployment.....")
	result, err := deploymentsClient.Create(deployment) // deployment yaml 객체 리턴
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	prompt()
	//=====================================================
	//=============Deployment Controller  생성 ================
	fmt.Println("Updating deployment.....")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := deploymentsClient.Get(metaName, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Faild to get latest version of Deployement : %v", getErr))
		}

		result.Spec.Replicas = pointer.Int32Ptr(3)
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13"
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment.....")
	//=====================================================

	//=============Deployment Controller  읽기 ================
	prompt()
	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Println("==================")
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
		prettyPrint(d)
	}

	//=============Deployment Controller  삭제 ================
	prompt()
	fmt.Println("Deleting deployment.....")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(metaName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")

}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	print(string(s))
	return string(s)
}
