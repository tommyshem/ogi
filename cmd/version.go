package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// bump version for bug fixes and features
// not-backward-compatible.new-feature.bug fixes
var Version = "0.1.2"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of OGI.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OGI v%s\n", Version)
	},
}

// init registers the version command with the root command.
func init() {
	RootCmd.AddCommand(versionCmd)
}
