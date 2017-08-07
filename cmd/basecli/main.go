package main

import (
	"os"

	"github.com/spf13/cobra"

	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/tmlibs/cli"

	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/auto"
	"github.com/tendermint/basecoin/client/commands/proxy"
	"github.com/tendermint/basecoin/client/commands/query"
	rpccmd "github.com/tendermint/basecoin/client/commands/rpc"
	"github.com/tendermint/basecoin/client/commands/seeds"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	authcmd "github.com/tendermint/basecoin/modules/auth/commands"
	basecmd "github.com/tendermint/basecoin/modules/base/commands"
	coincmd "github.com/tendermint/basecoin/modules/coin/commands"
	feecmd "github.com/tendermint/basecoin/modules/fee/commands"
	ibccmd "github.com/tendermint/basecoin/modules/ibc/commands"
	noncecmd "github.com/tendermint/basecoin/modules/nonce/commands"
	rolecmd "github.com/tendermint/basecoin/modules/roles/commands"
)

// BaseCli - main basecoin client command
var BaseCli = &cobra.Command{
	Use:   "basecli",
	Short: "Light client for Tendermint",
	Long: `Basecli is a certifying light client for the basecoin abci app.

It leverages the power of the tendermint consensus algorithm get full
cryptographic proof of all queries while only syncing a fraction of the
block headers.`,
}

func main() {
	commands.AddBasicFlags(BaseCli)

	// Prepare queries
	query.RootCmd.AddCommand(
		// These are default parsers, but optional in your app (you can remove key)
		query.TxQueryCmd,
		query.KeyQueryCmd,
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,
		rolecmd.RoleQueryCmd,
		ibccmd.IBCQueryCmd,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{
		feecmd.FeeWrapper{},
		rolecmd.RoleWrapper{},
		noncecmd.NonceWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	// you will always want this for the base send command
	txcmd.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		coincmd.SendTxCmd,
		coincmd.CreditTxCmd,
		// this enables creating roles
		rolecmd.CreateRoleTxCmd,
		// these are for handling ibc
		ibccmd.RegisterChainTxCmd,
		ibccmd.UpdateChainTxCmd,
		ibccmd.PostPacketTxCmd,
	)

	// Set up the various commands to use
	BaseCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		keycmd.RootCmd,
		seeds.RootCmd,
		rpccmd.RootCmd,
		query.RootCmd,
		txcmd.RootCmd,
		proxy.RootCmd,
		commands.VersionCmd,
		auto.AutoCompleteCmd,
	)

	cmd := cli.PrepareMainCmd(BaseCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	cmd.Execute()
}
