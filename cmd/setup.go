package cmd

import (
	"fmt"
	"os"

	"github.com/franciscocunha55/chaos_engineering/pkg/chaos"
	"github.com/franciscocunha55/chaos_engineering/pkg/k8s"
	"github.com/spf13/cobra"
)

var setupNamespace string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup test environment (namespace + nginx deployment)",
	Long:  `Creates a namespace and deploys a test nginx application for chaos engineering tests.`,
	Run:   runSetup,
}

func init() {
	setupCmd.Flags().StringVarP(&setupNamespace, "namespace", "n", "chaos-engineering-test", "Namespace to create")
	rootCommand.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	fmt.Printf("Setting up chaos engineering environment in namespace: %s\n", setupNamespace)

	clientSet, err := k8s.GetClientSet()
	if err != nil {
		fmt.Printf("Error creating clientset: %v\n", err)
		os.Exit(1)
	}

	namespacesList := chaos.ListNamespaces(clientSet)

	fmt.Printf("Creating namespace %s...\n", setupNamespace)
	chaos.CreateNamespace(clientSet, setupNamespace, namespacesList)

	fmt.Printf("Creating nginx deployment in %s...\n", setupNamespace)
	chaos.CreateDeployment(clientSet, setupNamespace, "chaos-engineering-nginx")

	fmt.Println("\n✓ Setup completed successfully!")
	fmt.Printf("You can now run: chaos kill -n %s -l app=chaos-engineering-nginx --random\n", setupNamespace)
}
