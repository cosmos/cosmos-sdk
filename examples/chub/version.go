package main

import "github.com/spf13/cobra"

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		RunE:  todoNotImplemented,
	}
)
