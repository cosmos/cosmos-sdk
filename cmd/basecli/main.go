package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/abci/version"
	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/basecoin/commands"
	"github.com/tendermint/basecoin/commands/proofs"
	"github.com/tendermint/basecoin/commands/proxy"
	rpccmd "github.com/tendermint/basecoin/commands/rpc"
	"github.com/tendermint/basecoin/commands/seeds"
	"github.com/tendermint/basecoin/commands/txs"
	"github.com/tendermint/tmlibs/cli"

	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	authcmd "github.com/tendermint/basecoin/modules/auth/commands"
	basecmd "github.com/tendermint/basecoin/modules/base/commands"
	coincmd "github.com/tendermint/basecoin/modules/coin/commands"
	feecmd "github.com/tendermint/basecoin/modules/fee/commands"
	noncecmd "github.com/tendermint/basecoin/modules/nonce/commands"
)

// BaseCli - main basecoin client command
var BaseCli = &cobra.Command{
	Use:   "basecli",
	Short: "Light client for tendermint",
	Long: `Basecli is an version of tmcli including custom logic to
present a nice (not raw hex) interface to the basecoin blockchain structure.

This is a useful tool, but also serves to demonstrate how one can configure
tmcli to work for any custom abci app.
`,
}

// VersionCmd - command to show the application version
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func main() {
	commands.AddBasicFlags(BaseCli)

	// Prepare queries
	proofs.RootCmd.AddCommand(
		// These are default parsers, but optional in your app (you can remove key)
		proofs.TxCmd,
		proofs.KeyCmd,
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,
	)

	// set up the middleware
	bcmd.Middleware = bcmd.Wrappers{
		feecmd.FeeWrapper{},
		noncecmd.NonceWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	bcmd.Middleware.Register(txs.RootCmd.PersistentFlags())

	// you will always want this for the base send command
	proofs.TxPresenters.Register("base", bcmd.BaseTxPresenter{})
	txs.RootCmd.AddCommand(
		// This is the default transaction, optional in your app
		coincmd.SendTxCmd,
	)

	// Set up the various commands to use
	BaseCli.AddCommand(
		commands.InitCmd,
		commands.ResetCmd,
		keycmd.RootCmd,
		seeds.RootCmd,
		rpccmd.RootCmd,
		proofs.RootCmd,
		txs.RootCmd,
		proxy.RootCmd,
		VersionCmd,
		bcmd.AutoCompleteCmd,
	)

	cmd := cli.PrepareMainCmd(BaseCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	cmd.Execute()
}
