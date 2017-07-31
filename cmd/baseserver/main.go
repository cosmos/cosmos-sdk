package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/client/commands"
	rest "github.com/tendermint/basecoin/client/rest"
	"github.com/tendermint/tmlibs/cli"
)

var srvCli = &cobra.Command{
	Use:   "baseserver",
	Short: "Light REST client for tendermint",
	Long:  `Baseserver presents  a nice (not raw hex) interface to the basecoin blockchain structure.`,
}

func main() {
	commands.AddBasicFlags(srvCli)

	srvCli.AddCommand(
		commands.InitCmd,
		rest.ServeCmd,
	)

	// TODO: Decide whether to use $HOME/.basecli for compatibility
	// or just use $HOME/.baseserver?
	cmd := cli.PrepareMainCmd(srvCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
