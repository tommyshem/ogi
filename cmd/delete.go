package cmd

import (
	"github.com/spf13/cobra"
)

// TODO implement the delete database file
var deleteDB = &cobra.Command{
	Use:   "delete",
	Short: "delete the local offline database file on your system.",
	Run: func(cmd *cobra.Command, args []string) {
		//	filePath := Location()
	},
}

// init registers the list command with the root command and sets up flags on
// the list command.
func init() {
	RootCmd.AddCommand(deleteDB)
	listCmd.Flags().BoolVarP(&raw, "raw", "r", false, "Show the raw JSON for these issues")
	listCmd.Flags().StringVarP(&state, "state", "s", "open", "List issues by their state <all, closed, open>")
}
