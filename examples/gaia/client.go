package main

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/commits"
	"github.com/cosmos/cosmos-sdk/client/commands/keys"
	"github.com/cosmos/cosmos-sdk/client/commands/proxy"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	rpccmd "github.com/cosmos/cosmos-sdk/client/commands/rpc"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	authcmd "github.com/cosmos/cosmos-sdk/modules/auth/commands"
	basecmd "github.com/cosmos/cosmos-sdk/modules/base/commands"
	coincmd "github.com/cosmos/cosmos-sdk/modules/coin/commands"
	feecmd "github.com/cosmos/cosmos-sdk/modules/fee/commands"
	ibccmd "github.com/cosmos/cosmos-sdk/modules/ibc/commands"
	noncecmd "github.com/cosmos/cosmos-sdk/modules/nonce/commands"
	rolecmd "github.com/cosmos/cosmos-sdk/modules/roles/commands"

	stakecmd "github.com/cosmos/gaia/modules/stake/commands"
)

// clientCmd is the entry point for this binary
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Gaia light client",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func prepareClientCommands() {
	commands.AddBasicFlags(clientCmd)

	// Prepare queries
	query.RootCmd.AddCommand(
		// These are default parsers, but optional in your app (you can remove key)
		query.TxQueryCmd,
		query.KeyQueryCmd,
		coincmd.AccountQueryCmd,
		noncecmd.NonceQueryCmd,
		rolecmd.RoleQueryCmd,
		ibccmd.IBCQueryCmd,

		//stakecmd.CmdQueryValidator,
		stakecmd.CmdQueryCandidates,
		stakecmd.CmdQueryCandidate,
		stakecmd.CmdQueryDelegatorBond,
		stakecmd.CmdQueryDelegatorCandidates,
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

		stakecmd.CmdDeclareCandidacy,
		stakecmd.CmdEditCandidacy,
		stakecmd.CmdDelegate,
		stakecmd.CmdUnbond,
	)

	clientCmd.AddCommand(
		proxy.RootCmd,
		lineBreak,

		txcmd.RootCmd,
		query.RootCmd,
		rpccmd.RootCmd,
		lineBreak,

		keys.RootCmd,
		commands.InitCmd,
		commands.ResetCmd,
		commits.RootCmd,
		lineBreak,
	)

}
