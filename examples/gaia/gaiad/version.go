package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/version"
)

var (
	// VersionCmd prints out the current sdk version
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		Run:   doVersionCmd,
	}
)

func doVersionCmd(cmd *cobra.Command, args []string) {
	v := version.Version
	if version.GitCommit != "" {
		v = v + " " + version.GitCommit
	}
	fmt.Println(v)
}
