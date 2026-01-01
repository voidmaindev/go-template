package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, commit hash, and build date of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GoTemplate API\n")
		fmt.Printf("  Version:    %s\n", Version)
		fmt.Printf("  Commit:     %s\n", Commit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
