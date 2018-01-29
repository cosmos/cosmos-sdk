package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/version"
)

var rootCmd = &cobra.Command{
	Use:   "basecoind",
	Short: "The basecoin daemon.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info for basecoind.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func main() {
	rootCmd.AddCommand(
		versionCmd,
	)

	cmd := cli.PrepareMainCmd(rootCmd, "BC", os.ExpandEnv("$HOME/.basecoind"))
	cmd.Execute()
}
