package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	metrics "github.com/franciscocunha55/chaos_engineering/pkg/metrics"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


func PerformChaosTest(clientSet *kubernetes.Clientset, namespace string) {
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
			metrics.ChaosPodsDeletedCounter.WithLabelValues(namespace).Inc()
		} else {
			fmt.Println("No pods found to delete")
		}
}

func DeletePod(clientSet *kubernetes.Clientset, pod apiv1.Pod, dryRun bool) error{
	if dryRun{
		fmt.Printf("[DRY-RUN] Would delete pod %s in namespace %s\n", pod.Name, pod.Namespace)
		return nil
	}
	err := clientSet.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err != nil{
		return fmt.Errorf("failed to delete pod %s: %w", pod.Name, err)
	}
	fmt.Printf("Successfully deleted pod: %s\n", pod.Name)
	return nil
}
