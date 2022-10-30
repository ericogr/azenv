package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "azenv",
	Short: "AzureDevOps Environment Management",
	Long: fmt.Sprintf(`
This tool can manage Azure DevOps environments (for now, only Kubernetes is supported)
Version: %s
Git Tag: %s
Build Date: %s
`, Version, GitTag, BuildDate),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("type", "t", "kubernetes", "Environment resource type (for now, only Kubernetes is supported)")
	rootCmd.PersistentFlags().String("pat", "", "AzureDevOps Personal Access Token (PAT)")
	err := rootCmd.MarkPersistentFlagRequired("pat")
	if err != nil {
		fmt.Println(err.Error())
	}
}
