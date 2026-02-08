package cmd

import (
	"fmt"
	"os"

	"github.com/franciscocunha55/chaos_engineering/pkg/chaos"
	"github.com/franciscocunha55/chaos_engineering/pkg/k8s"
	"github.com/franciscocunha55/chaos_engineering/pkg/metrics"
	"github.com/spf13/cobra"
)

var (
    namespace     string
    labelSelector string
    percentage    int
    random        bool
    dryRun        bool
)

var killCmd = &cobra.Command{
    Use:   "kill",
    Short: "Kill pods based on selection criteria",
    Run:   runKill,
}

func init(){
	killCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to target (required)")
    killCmd.Flags().StringVarP(&labelSelector, "label", "l", "", "Label selector (required)")
    killCmd.Flags().IntVarP(&percentage, "percentage", "p", 0, "Percentage of pods to kill (0-100)")
    killCmd.Flags().BoolVarP(&random, "random", "r", false, "Kill one random pod")
    killCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Dry run mode")
    
    killCmd.MarkFlagRequired("namespace")
    killCmd.MarkFlagRequired("label")
	rootCommand.AddCommand(killCmd)
}

func runKill(cmd *cobra.Command, args []string){
	if !random && percentage == 0 {
        fmt.Println("Error: must specify --random or --percentage")
        os.Exit(1)
    }

	if random && percentage > 0 {
        fmt.Println("Error: cannot use both --random and --percentage")
        os.Exit(1)
    }

	clientSet, err := k8s.GetClientSet()
    if err != nil {
        fmt.Printf("Error creating clientset: %v\n", err)
        os.Exit(1)
    }

	opts := chaos.PodSelectionOptions{
        Namespace:     namespace,
        LabelSelector: labelSelector,
        Percentage:    percentage,
        Random:        random,
        DryRun:        dryRun,
    }

	pods, err := chaos.ListPods(clientSet, opts)
    if err != nil {
        fmt.Printf("Error listing pods: %v\n", err)
        os.Exit(1)
    }
    
    if len(pods) == 0 {
        fmt.Println("No pods found matching criteria")
        return
    }
	
	fmt.Printf("Found %d pods matching label %s\n", len(pods), labelSelector)

	selectedPods := chaos.SelectPodsForChaos(pods, opts)
	if len(selectedPods) == 0 {
        fmt.Println("No pods selected for chaos")
        return
    }
    
    fmt.Printf("Selected %d pods for deletion\n", len(selectedPods))

	for _, pod := range selectedPods {
        err := chaos.DeletePod(clientSet, pod, opts.DryRun)
        if err != nil {
            fmt.Printf("Error deleting pod %s: %v\n", pod.Name, err)
            continue
        }
        
        if !opts.DryRun {
            metrics.ChaosPodsDeletedCounter.WithLabelValues(namespace).Inc()
        }
    }
    
    fmt.Println("Chaos operation completed")
}