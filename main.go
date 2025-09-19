package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("could not determine home directory")
	}
	
	kubeconfig := filepath.Join(homeDir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//clientset to talk to core resources like Pods
	clientSet, err :=kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespacesList, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found %d namespaces:\n", len(namespacesList.Items))
	for _, ns := range namespacesList.Items {
		fmt.Printf("- %s\n", ns.Name)
	}

	chaosEngineeringNamespaceName := "chaos-engineering-test"

	namespaceExists := false
	for _, ns := range namespacesList.Items {
		if ns.Name == chaosEngineeringNamespaceName {
			namespaceExists = true
			break
		}
	}
	if namespaceExists {	
		fmt.Printf("Namespace %s already exists, skipping creation \n", chaosEngineeringNamespaceName)
	} else {
		fmt.Printf("Creating namespace %s\n", chaosEngineeringNamespaceName)
		namespaceInfo := &v1.Namespace{
            ObjectMeta: metav1.ObjectMeta{
                Name: chaosEngineeringNamespaceName,
            },
        }
		_, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), namespaceInfo, metav1.CreateOptions{})
		if err != nil {
			panic(err.Error())
		}
	}

	podsChaosEngineeringNamespace, err:= clientSet.CoreV1().Pods("chaos-engineering-test").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=postgresql",
	})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found %d pods in %s namespace:\n", len(podsChaosEngineeringNamespace.Items), chaosEngineeringNamespaceName)
	for _, pod := range podsChaosEngineeringNamespace.Items {
		fmt.Printf("- [%s] %s\n", pod.Namespace, pod.Name)
	}
}
