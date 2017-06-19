package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/commands/proxy"
	"github.com/tendermint/light-client/commands/seeds"
	"github.com/tendermint/light-client/commands/txs"
	"github.com/tendermint/tmlibs/cli"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	bcount "github.com/tendermint/basecoin/docs/guide/counter/cmd/countercli/commands"
)

// BaseCli represents the base command when called without any subcommands
var BaseCli = &cobra.Command{
	Use:   "countercli",
	Short: "Light client for tendermint",
	Long: `Basecli is an version of tmcli including custom logic to
present a nice (not raw hex) interface to the basecoin blockchain structure.

This is a useful tool, but also serves to demonstrate how one can configure
tmcli to work for any custom abci app.
`,
}

func main() {
	commands.AddBasicFlags(BaseCli)

	// Prepare queries
	proofs.RootCmd.AddCommand(
		// These are default parsers, optional in your app
		proofs.TxCmd,
		proofs.KeyCmd,
		bcmd.AccountQueryCmd,

		// XXX IMPORTANT: here is how you add custom query commands in your app
		bcount.CounterQueryCmd,
	)

	// Prepare transactions
	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	txs.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		bcmd.SendTxCmd,

		// XXX IMPORTANT: here is how you add custom tx construction for your app
		bcount.CounterTxCmd,
	)

	// Set up the various commands to use
	BaseCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		keycmd.RootCmd,
		seeds.RootCmd,
		proofs.RootCmd,
		txs.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(BaseCli, "CTL", os.ExpandEnv("$HOME/.countercli"))
	cmd.Execute()
}
