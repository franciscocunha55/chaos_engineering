package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	podsDefaultNamespace, err:= clientSet.CoreV1().Pods("postgres").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found %d pods:\n", len(podsDefaultNamespace.Items))
	for _, pod := range podsDefaultNamespace.Items {
		fmt.Printf("- [%s] %s\n", pod.Namespace, pod.Name)
	}
}
