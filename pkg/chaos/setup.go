package chaos

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"k8s.io/client-go/kubernetes"
)

func ListNamespaces(clientSet *kubernetes.Clientset) *apiv1.NamespaceList{
	namespacesList, err := clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found %d namespaces:\n", len(namespacesList.Items))
	for _, ns := range namespacesList.Items {
		fmt.Printf("- %s\n", ns.Name)
	}
	return namespacesList
}

func CreateNamespace(clientSet *kubernetes.Clientset, namespaceName string, namespacesList *apiv1.NamespaceList) {
	namespaceExists := false
	for _, ns := range namespacesList.Items {
		if ns.Name == namespaceName {
			namespaceExists = true
			break
		}
	}
	if namespaceExists {	
		fmt.Printf("Namespace %s already exists, skipping creation \n", namespaceName)
	} else {
		fmt.Printf("Creating namespace %s\n", namespaceName)
		namespaceInfo := &apiv1.Namespace{
            ObjectMeta: metav1.ObjectMeta{
                Name: namespaceName,
            },
        }
		_, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), namespaceInfo, metav1.CreateOptions{})
		if err != nil {
			panic(err.Error())
		}
	}
}

func CreateDeployment(clientSet *kubernetes.Clientset, namespaceName string, deploymentName string) {
    deploymentsList, err := clientSet.AppsV1().Deployments(namespaceName).List(context.TODO(), metav1.ListOptions{})
    if err != nil {
        panic(err.Error())
    }
    
    deploymentNginxExists := false
    for _, deploy := range deploymentsList.Items {
        if deploy.Name == deploymentName {
            deploymentNginxExists = true
            break
        }
    }
    
    if deploymentNginxExists {
        fmt.Printf("Deployment chaos-engineering-nginx already exists in %s namespace, skipping creation \n", namespaceName)
    } else {
        fmt.Printf("Creating deployment chaos-engineering-nginx in %s namespace \n", namespaceName)
        deploymentNginxDeclaration := &appsv1.Deployment{
            ObjectMeta: metav1.ObjectMeta{
                Name: "chaos-engineering-nginx",
            },
            Spec: appsv1.DeploymentSpec{
                Replicas: ptr.To[int32](3),
                Selector: &metav1.LabelSelector{
                    MatchLabels: map[string]string{
                        "app": deploymentName,
                    },
                },
                Template: apiv1.PodTemplateSpec{
                    ObjectMeta: metav1.ObjectMeta{
                        Labels: map[string]string{
                            "app": deploymentName,
                        },
                    },
                    Spec: apiv1.PodSpec{
                        Containers: []apiv1.Container{
                            {
                                Name:  "nginx",
                                Image: "nginx:1.12",
                                Ports: []apiv1.ContainerPort{
                                    {
                                        Name:          "http",
                                        Protocol:      apiv1.ProtocolTCP,
                                        ContainerPort: 80,
                                    },
                                },
                            },
							{
                                Name:  "stress-ng",
                                Image: "ubuntu:22.04",
                                Command: []string{"sh"},
                                Args: []string{"-c", "apt-get update && apt-get install -y stress-ng && stress-ng --cpu 3 --timeout 600"},
                            },
                        },
                    },
                },
            },
        }
        deploymentNginx, err := clientSet.AppsV1().Deployments(namespaceName).Create(context.TODO(), deploymentNginxDeclaration, metav1.CreateOptions{})
        if err != nil {
            panic(err.Error())
        }
        fmt.Printf("Created deployment %q.\n", deploymentNginx.Name)
		time.Sleep(10 * time.Second) // Wait for pods to be created
    }
}