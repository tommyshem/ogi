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
	Long:  `OGI let's you download issues from a GitHub repo's to be made available offline.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
