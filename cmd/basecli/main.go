package main

import (
	"os"

	"github.com/spf13/cobra"
	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/commands/seeds"
	"github.com/tendermint/light-client/commands/txs"
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

func init() {
	commands.AddBasicFlags(BaseCli)

	// set up the various commands to use
	BaseCli.AddCommand(keycmd.RootCmd)
	BaseCli.AddCommand(commands.InitCmd)
	BaseCli.AddCommand(seeds.RootCmd)
	proofs.StatePresenters.Register("account", AccountPresenter{})
	proofs.TxPresenters.Register("base", BaseTxPresenter{})
	BaseCli.AddCommand(proofs.RootCmd)
	txs.Register("send", SendTxMaker{})
	BaseCli.AddCommand(txs.RootCmd)
}

func main() {
	keycmd.PrepareMainCmd(BaseCli, "BC", os.ExpandEnv("$HOME/.basecli"))
	BaseCli.Execute()
	// err := BaseCli.Execute()
	// if err != nil {
	// 	fmt.Printf("%+v\n", err)
	// }
}
