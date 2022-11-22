package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var logger *log.Logger
var rootCmd = &cobra.Command{
	PreRun: toggleDebug,
	Use:    "azenv",
	Short:  "AzureDevOps Environment Management",
	Long: `This tool can manage Azure DevOps environments (for now, only Kubernetes is supported)

Example:
azenv create kubernetes \
    --pat your-azuredevops-pat \
    --name new-test-environment \
    --project totvsappfoundation/TOTVSApps \
    --service-account new-test-namespace/test-sa \
    --service-connection new-test-service-connection \
    --namespace-label label1=value1 \
    --namespace-label label2=value2 \
    --show-kubeconfig=false
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no arguments")
		}

		return nil
	},
}

func toggleDebug(cmd *cobra.Command, args []string) {
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		logger.Fatal(err.Error())
	}

	if quiet {
		logger = log.New(io.Discard, "", 0)
	} else {
		logger = log.New(os.Stderr, "[azenv] ", log.LstdFlags)
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("quiet", false, "Only show output when errors are found")
}
