// Package main is the entry point for the aix CLI.
package main

import (
	"os"

	"github.com/thoreinstein/aix/cmd/aix/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
