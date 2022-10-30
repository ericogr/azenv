package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "azenv",
	Short: "AzureDevOps Environment Management",
	Long: fmt.Sprintf(`This tool can manage Azure DevOps environments (for now, only Kubernetes is supported)
v%s(%s) - %s

Example:
azenv create kubernetes \
    --pat your-azuredevops-pat \
    create kubernetes \
    --name new-test-environment \
    --project totvsappfoundation/TOTVSApps \
    --service-account new-test-namespace/test-sa \
    --service-connection new-test-service-connection \
    --namespace-label label1=value1 \
    --namespace-label label2=value2 \
    --show-kubeconfig true
`, Version, GitTag, BuildDate),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("pat", "", "[required] AzureDevOps Personal Access Token (PAT)")
	err := rootCmd.MarkPersistentFlagRequired("pat")
	if err != nil {
		fmt.Println(err.Error())
	}
}
