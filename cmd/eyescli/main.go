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
	etccmd "github.com/tendermint/basecoin/modules/eyes/commands"
)

// EyesCli - main basecoin client command
var EyesCli = &cobra.Command{
	Use:   "eyescli",
	Short: "Light client for Tendermint",
	Long:  `EyesCli is the light client for a merkle key-value store (eyes)`,
}

func main() {
	commands.AddBasicFlags(EyesCli)

	// Prepare queries
	query.RootCmd.AddCommand(
		// These are default parsers, but optional in your app (you can remove key)
		query.TxQueryCmd,
		query.KeyQueryCmd,
		// this is our custom parser
		etccmd.EtcQueryCmd,
	)

	// no middleware wrapers
	txcmd.Middleware = txcmd.Wrappers{}
	// txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	// just the etc commands
	txcmd.RootCmd.AddCommand(
		etccmd.SetTxCmd,
		etccmd.RemoveTxCmd,
	)

	// Set up the various commands to use
	EyesCli.AddCommand(
		// we use out own init command to not require address arg
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

	cmd := cli.PrepareMainCmd(EyesCli, "EYE", os.ExpandEnv("$HOME/.eyescli"))
	cmd.Execute()
}
