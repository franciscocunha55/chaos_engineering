package cmd

import "github.com/spf13/cobra"


var rootCommand = &cobra.Command{
	Use: "chaos",
	Short: "chaos is an app created to test reliability of kubernetes cluster",
}

	
func Execute() error{
	return rootCommand.Execute()
}


