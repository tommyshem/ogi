package main

import "github.com/tommyshem/ogi/cmd"

// main is the entry point of the application. It executes the root command
// which sets up all the available subcommands and flags for the CLI tool.
func main() {
	cmd.Execute()
}
