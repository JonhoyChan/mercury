package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
)

func Execute() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				fmt.Println(err.Error())
			} else {
				fmt.Println(r)
			}
		}
	}()

	ensureLargeNumOpenFiles()

	root, shutdown := NewMercuryCommand(StandardRepoPath())
	// If the subcommand hits an error, don't show usage or the error, since we'll show
	// the error message below, on our own. Usage is still shown if the subcommand
	// is missing command-line arguments.
	root.SilenceUsage = true
	root.SilenceErrors = true
	// Execute the subcommand
	if err := root.Execute(); err != nil {
		ErrExit(os.Stderr, err)
	}

	<-shutdown()
}

// ErrExit writes an error to the given io.Writer & exits
func ErrExit(w io.Writer, err error) {
	fmt.Fprintln(w, color.New(color.FgRed).Sprintf(err.Error(), err))
	os.Exit(1)
}
