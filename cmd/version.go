package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the build version
var Version string = "0.0.0"

// GitTag is the git tag of the build
var GitTag string = ""

// BuildDate is the date when the build was created
var BuildDate string = ""

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	PreRun: toggleDebug,
	Use:    "version",
	Short:  "Version information",
	Long:   `Use this command to get version information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version %s (%s)\n", Version, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
