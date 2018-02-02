package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/commits"
	"github.com/cosmos/cosmos-sdk/client/commands/keys"
	"github.com/cosmos/cosmos-sdk/client/commands/proxy"
	"github.com/cosmos/cosmos-sdk/client/commands/query"

	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	bcount "github.com/cosmos/cosmos-sdk/examples/counter/cmd/countercli/commands"
	authcmd "github.com/cosmos/cosmos-sdk/modules/auth/commands"
	basecmd "github.com/cosmos/cosmos-sdk/modules/base/commands"
	coincmd "github.com/cosmos/cosmos-sdk/modules/coin/commands"
	feecmd "github.com/cosmos/cosmos-sdk/modules/fee/commands"
	noncecmd "github.com/cosmos/cosmos-sdk/modules/nonce/commands"
)

// CounterCli represents the base command when called without any subcommands
var CounterCli = &cobra.Command{
	Use:   "countercli",
	Short: "Example app built using the Cosmos SDK",
	Long: `Countercli is a demo app that includes custom logic to
present a formatted interface to a custom blockchain structure.

This is a useful tool and also serves to demonstrate how to configure
the Cosmos SDK to work for any custom ABCI app, see:

`,
}

func main() {
	commands.AddBasicFlags(CounterCli)

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
	CounterCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		commands.VersionCmd,
		keys.RootCmd,
		commits.RootCmd,
		query.RootCmd,
		txcmd.RootCmd,
		proxy.RootCmd,
	)

	cmd := cli.PrepareMainCmd(CounterCli, "CTL", os.ExpandEnv("$HOME/.countercli"))
	cmd.Execute()
}
