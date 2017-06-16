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
	bcount "github.com/tendermint/basecoin/cmd/countercli/commands"
)

// BaseCli represents the base command when called without any subcommands
var BaseCli = &cobra.Command{
	Use:   "basecli",
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
	pr := proofs.RootCmd
	// These are default parsers, but you optional in your app
	pr.AddCommand(proofs.TxCmd)
	pr.AddCommand(proofs.KeyCmd)
	pr.AddCommand(bcmd.AccountQueryCmd)

	// IMPORTANT: here is how you add custom query commands in your app
	pr.AddCommand(bcount.CounterQueryCmd)

	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	tr := txs.RootCmd
	tr.AddCommand(bcmd.SendTxCmd)

	// IMPORTANT: here is how you add custom tx construction for your app
	tr.AddCommand(bcount.CounterTxCmd)

	// Set up the various commands to use
	BaseCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		keycmd.RootCmd,
		seeds.RootCmd,
		pr,
		tr,
		proxy.RootCmd)

	cmd := cli.PrepareMainCmd(BaseCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	cmd.Execute()
}
