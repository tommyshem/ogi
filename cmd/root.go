package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// change storage backend
	//	storage "github.com/tommyshem/ogi/cmd/storage/nutsdb"
	storage "github.com/tommyshem/ogi/cmd/storage/bolt"
	//storage "github.com/tommyshem/ogi/cmd/bolt"
)

// global structs
var db *storage.Store

// var db *storage.NutsStore
var config *Config

var RootCmd = &cobra.Command{
	Use:   "(OGI) Offline GitHub Issues",
	Short: fmt.Sprintf("Offline GitHub Issues (v%s)", Version),
	Long:  `OGI let's you download issues from a GitHub repo to be made available offline.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config = LoadConfig()
		if config.Repo == "" {
			if len(args) == 0 {
				fmt.Println(`It looks like you haven't initialized OGI yet!

The first time you run OGI you should run "ogi fetch owner/repo".

This will fetch all of your issues for that repository. Future calls
to "fetch" won't require the "owner/repo" since we'll store a little
meta-data file in this repo to track that.`)
				os.Exit(-1)
			} else {
				config.SetFromArgs(args)
			}
		}

		s, err := storage.New(config.Owner, config.Repo)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		db = s
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
