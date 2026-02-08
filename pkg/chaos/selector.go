package chaos

import (
	"context"
	"fmt"
	"math/rand"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


type PodSelectionOptions struct{
	Namespace string
	LabelSelector string
	Percentage int
	Random bool
	DryRun bool
}

func ListPods(clientSet *kubernetes.Clientset, podOpts PodSelectionOptions)([]apiv1.Pod,error){
	podList, err := clientSet.CoreV1().Pods(podOpts.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: podOpts.LabelSelector,
	})
	if err != nil{
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return podList.Items, nil
}

func SelectPodsForChaos(pods []apiv1.Pod, podOpts PodSelectionOptions) []apiv1.Pod {
	if len(pods) == 0 {
		return []apiv1.Pod{}
	}

	if podOpts.Random {
		randomIndex := rand.Intn(len(pods))
		return []apiv1.Pod{pods[randomIndex]}
	}

	if podOpts.Percentage>0 {
		count := (len(pods) * podOpts.Percentage) / 100
		if count == 0 && len(pods) > 0 {
			count = 1
		}
		if count > len(pods) {
			count = len(pods)
		}
		rand.Shuffle(len(pods), func(i, j int) {
			pods[i], pods[j] = pods[j], pods[i]
		})
		return pods[:count]
	}

	return []apiv1.Pod{}
}