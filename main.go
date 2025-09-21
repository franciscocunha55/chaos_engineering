package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
)

var chaosPodsDeletedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "chaos_pods_deleted_total",
		Help: "Total number of pods deleted by chaos engineering tests",
	},
	[]string{"namespace"},
)


//clientset to talk to core resources like Pods
func getClientSet() (*kubernetes.Clientset, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("could not determine home directory")
	}
	kubeconfig := filepath.Join(homeDir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return kubernetes.NewForConfig(config)
}

func listNamespaces(clientSet *kubernetes.Clientset) *apiv1.NamespaceList{
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

func createNaemespace(clientSet *kubernetes.Clientset, namespaceName string, namespacesList *apiv1.NamespaceList) {
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

func createDeployment(clientSet *kubernetes.Clientset, namespaceName string, deploymentName string, ) {
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
                                Name:  "web",
                                Image: "nginx:1.12",
                                Ports: []apiv1.ContainerPort{
                                    {
                                        Name:          "http",
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
        deploymentNginx, err := clientSet.AppsV1().Deployments(namespaceName).Create(context.TODO(), deploymentNginxDeclaration, metav1.CreateOptions{})
        if err != nil {
            panic(err.Error())
        }
        fmt.Printf("Created deployment %q.\n", deploymentNginx.Name)
    }
}

func performChaosTest(clientSet *kubernetes.Clientset, namespace string) {
	podsNginxChaosEngineeringNamespace, err:= clientSet.CoreV1().Pods("chaos-engineering-test").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=chaos-engineering-nginx",
	})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Found %d pods in %s namespace:\n", len(podsNginxChaosEngineeringNamespace.Items), namespace)
	for _, pod := range podsNginxChaosEngineeringNamespace.Items {
		fmt.Printf("- [%s] %s\n", pod.Namespace, pod.Name)
	}

	if len(podsNginxChaosEngineeringNamespace.Items) > 0 {
			rand.Seed(time.Now().UnixNano())
			randomIndex := rand.Intn(len(podsNginxChaosEngineeringNamespace.Items))
			podToDelete := podsNginxChaosEngineeringNamespace.Items[randomIndex]
			fmt.Printf("Deleting pod %s in namespace %s\n", podToDelete.Name, podToDelete.Namespace)
			err := clientSet.CoreV1().Pods(podToDelete.Namespace).Delete(context.TODO(), podToDelete.Name, metav1.DeleteOptions{})
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("Successfully deleted pod: %s\n", podToDelete.Name)
			chaosPodsDeletedCounter.WithLabelValues(namespace).Inc()
		} else {
			fmt.Println("No pods found to delete")
		}
}


func main() {

	// Flags
	intervalBetweenChaosTest := flag.Int("interval", 60, "Interval between chaos tests in seconds")
	chaosEngineeringNamespaceName := flag.String("namespace", "chaos-engineering-test", "Namespace for chaos engineering tests")
	flag.Parse()
	fmt.Printf("Interval between chaos tests: %d seconds\n", *intervalBetweenChaosTest)

	prometheus.MustRegister(chaosPodsDeletedCounter)
	go func()  {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()


	clientSet, err := getClientSet()
	if err != nil {
		panic(err.Error())
	}

	namespacesList := listNamespaces(clientSet)

	createNaemespace(clientSet, *chaosEngineeringNamespaceName, namespacesList)

	createDeployment(clientSet, *chaosEngineeringNamespaceName, "chaos-engineering-nginx")

	fmt.Printf("Starting chaos tests every %d seconds...\n", *intervalBetweenChaosTest)

	//ticker is a clock that ticks at regular intervals
	ticker := time.NewTicker(time.Duration(*intervalBetweenChaosTest) * time.Second)
	//Ensures the ticker is stopped when the program exits
	defer ticker.Stop()

	performChaosTest(clientSet, *chaosEngineeringNamespaceName)

	// ticker.C is a channel that receives a signal every time the ticker ticks, Repeat until program is terminated
	for range ticker.C {
        performChaosTest(clientSet, *chaosEngineeringNamespaceName)
    }
}
