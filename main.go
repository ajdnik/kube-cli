package main

import (
	"fmt"
	"os"

	"github.com/ajdnik/kube-cli/commands"
	"github.com/ajdnik/kube-cli/version"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "kube-cli",
	Short: "Kubernetes deployment tool",
	Long: `Kubernetes deployment tool simplifies the DevOps workflow by automating 
container build, deployment configuration and container deployment steps.`,
	Version: version.GetVersion(),
}

func main() {
	cobra.OnInitialize()
	root.AddCommand(commands.UpdateCommand)
	root.AddCommand(commands.DeployCommand)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
