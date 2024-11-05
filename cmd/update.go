package cmd

import (
	"github.com/spf13/cobra"
)

var updateAll bool

// updateCmd update issues
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update offline issues to local ogi database",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// init registers the version command with the root command.
func init() {
	RootCmd.AddCommand(updateCmd)
	fetchCmd.Flags().BoolVarP(&updateAll, "all", "a", false, "Fetch issues from all the repo's stored in the offline database")
}
