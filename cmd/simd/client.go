package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/simapp"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

var (
	encodingConfig = simapp.MakeEncodingConfig()
	initClientCtx  = client.Context{}.
			WithJSONMarshaler(encodingConfig.Marshaler).
			WithTxGenerator(encodingConfig.TxGenerator).
			WithCodec(encodingConfig.Amino).
			WithInput(os.Stdin).
			WithAccountRetriever(types.NewAccountRetriever(encodingConfig.Marshaler))
)

func init() {
	authclient.Codec = encodingConfig.Marshaler
}

// TODO: setup keybase, viper object, etc. to be passed into
// the below functions and eliminate global vars, like we do
// with the cdc
func addClientCommands(rootClientCmd *cobra.Command) context.Context {
	// Add --chain-id to persistent flags and mark it required
	rootClientCmd.PersistentFlags().String(flags.FlagChainID, "", "network chain ID")

	// Construct Root Command
	rootClientCmd.AddCommand(
		rpc.StatusCommand(),
		queryCmd(),
		txCmd(),
		keys.Commands(),
		flags.NewCompletionCmd(rootClientCmd, true),
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})

	return ctx
}

func queryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return client.SetCmdClientContextHandler(initClientCtx, cmd)
		},
		RunE: client.ValidateCmd,
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(encodingConfig.Amino),
		rpc.ValidatorCommand(encodingConfig.Amino),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(encodingConfig.Amino),
		authcmd.QueryTxCmd(encodingConfig.Amino),
	)

	simapp.ModuleBasics.AddQueryCommands(queryCmd, initClientCtx)

	return queryCmd
}

func txCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return client.SetCmdClientContextHandler(initClientCtx, cmd)
		},
		RunE: client.ValidateCmd,
	}

	txCmd.AddCommand(
		bankcmd.NewSendTxCmd(),
		flags.LineBreak,
		authcmd.GetSignCommand(initClientCtx),
		authcmd.GetSignBatchCommand(encodingConfig.Amino),
		authcmd.GetMultiSignCommand(initClientCtx),
		authcmd.GetValidateSignaturesCommand(initClientCtx),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(initClientCtx),
		authcmd.GetEncodeCommand(initClientCtx),
		authcmd.GetDecodeCommand(initClientCtx),
		flags.LineBreak,
	)

	simapp.ModuleBasics.AddTxCommands(txCmd, initClientCtx)

	return txCmd
}
