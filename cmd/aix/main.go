package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/thoreinstein/aix/cmd/aix/commands"
	aixerrors "github.com/thoreinstein/aix/internal/errors"
)

func main() {
	if err := commands.Execute(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	var exitErr *aixerrors.ExitError
	if errors.As(err, &exitErr) {
		prefix := "Error:"
		if exitErr.Code == aixerrors.ExitSystem {
			prefix = "Internal error:"
		}

		// Print formatted error
		printFormatted(prefix, exitErr.Error())

		if exitErr.Suggestion != "" {
			fmt.Fprintf(os.Stderr, "%s\n", exitErr.Suggestion)
		}
		os.Exit(exitErr.Code)
	}

	// Generic error (likely Cobra flag error)
	printFormatted("Error:", err.Error())
	os.Exit(aixerrors.ExitUser)
}

func printFormatted(prefix, msg string) {
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	fmt.Fprintf(os.Stderr, "%s %s\n", red(prefix), bold(msg))
}
