package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tommyshem/ogi/cmd/issue"
)

var state string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists issues for the repo.",
	Run: func(cmd *cobra.Command, args []string) {
		var issues []issue.Issue
		var err error
		// state
		switch state {
		case "all":
			issues, err = db.All()
		case "closed":
			println("Closed Issues Only")
			issues, err = db.AllByState(state)
		case "open":
			println("Opened Issues Only")
			issues, err = db.AllByState(state)
		default:
			issues, err = db.All()
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		// raw flag
		if raw {
			b, err := json.MarshalIndent(issues, "", "  ")
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			fmt.Print(string(b))
		} else {
			for _, issue := range issues {
				fmt.Print(issue.FmtTitle())
			}
			fmt.Printf("\n=== (%d) Issues ===\n", len(issues))
		}
	},
}

// init registers the list command with the root command and sets up flags on
// the list command.
func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&raw, "raw", "r", false, "Show the raw JSON for these issues")
	listCmd.Flags().StringVarP(&state, "state", "s", "open", "List issues by their state <all, closed, open>")
}
