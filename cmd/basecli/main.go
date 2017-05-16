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

	//initialize proofs and txs
	proofs.StatePresenters.Register("account", AccountPresenter{})
	proofs.StatePresenters.Register("counter", CounterPresenter{})
	proofs.TxPresenters.Register("base", BaseTxPresenter{})

	txs.Register("send", SendTxMaker{})
	txs.Register("counter", CounterTxMaker{})

	// set up the various commands to use
	BaseCli.AddCommand(
		keycmd.RootCmd,
		commands.InitCmd,
		seeds.RootCmd,
		proofs.RootCmd,
		txs.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(BaseCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	cmd.Execute()
}
