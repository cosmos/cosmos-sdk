package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/basecoin/commands"
	"github.com/tendermint/basecoin/commands/proofs"
	"github.com/tendermint/basecoin/commands/proxy"
	"github.com/tendermint/basecoin/commands/seeds"
	"github.com/tendermint/basecoin/commands/txs"
	"github.com/tendermint/tmlibs/cli"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	bcount "github.com/tendermint/basecoin/docs/guide/counter/cmd/countercli/commands"
	authcmd "github.com/tendermint/basecoin/modules/auth/commands"
	basecmd "github.com/tendermint/basecoin/modules/base/commands"
	coincmd "github.com/tendermint/basecoin/modules/coin/commands"
	feecmd "github.com/tendermint/basecoin/modules/fee/commands"
	noncecmd "github.com/tendermint/basecoin/modules/nonce/commands"
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
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,

		// XXX IMPORTANT: here is how you add custom query commands in your app
		bcount.CounterQueryCmd,
	)

	// set up the middleware
	bcmd.Middleware = bcmd.Wrappers{
		feecmd.FeeWrapper{},
		noncecmd.NonceWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	bcmd.Middleware.Register(txs.RootCmd.PersistentFlags())

	// Prepare transactions
	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	txs.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		coincmd.SendTxCmd,

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
