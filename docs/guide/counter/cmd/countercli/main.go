package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/proxy"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/client/commands/seeds"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	bcount "github.com/cosmos/cosmos-sdk/docs/guide/counter/cmd/countercli/commands"
	authcmd "github.com/cosmos/cosmos-sdk/modules/auth/commands"
	basecmd "github.com/cosmos/cosmos-sdk/modules/base/commands"
	coincmd "github.com/cosmos/cosmos-sdk/modules/coin/commands"
	feecmd "github.com/cosmos/cosmos-sdk/modules/fee/commands"
	noncecmd "github.com/cosmos/cosmos-sdk/modules/nonce/commands"
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
	query.RootCmd.AddCommand(
		// These are default parsers, optional in your app
		query.TxQueryCmd,
		query.KeyQueryCmd,
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,

		// XXX IMPORTANT: here is how you add custom query commands in your app
		bcount.CounterQueryCmd,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{
		feecmd.FeeWrapper{},
		noncecmd.NonceWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	// Prepare transactions
	txcmd.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		coincmd.SendTxCmd,

		// XXX IMPORTANT: here is how you add custom tx construction for your app
		bcount.CounterTxCmd,
	)

	// Set up the various commands to use
	BaseCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		commands.VersionCmd,
		keycmd.RootCmd,
		seeds.RootCmd,
		query.RootCmd,
		txcmd.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(BaseCli, "CTL", os.ExpandEnv("$HOME/.countercli"))
	cmd.Execute()
}
