package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tommyshem/ogi/cmd/storage/bolt"
)

var delete bool

// TODO implement the delete database file
var deleteDB = &cobra.Command{
	Use:   "delete",
	Short: "Delete the local ogi offline database file on your system.",
	Run: func(cmd *cobra.Command, args []string) {
		err := os.Remove(bolt.Location())
		if err != nil {
			log.Fatal(err)
		}
	},
}

// init registers the delete command with the root command and sets up flags on
// the list command.
func init() {
	RootCmd.AddCommand(deleteDB)
	listCmd.Flags().BoolVarP(&delete, "delete", "d", false, "Delete the offline issues database from the local machine")
}
