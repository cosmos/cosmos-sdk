package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	rootCmd := &cobra.Command{
		Use:   "simcli",
		Short: "Command line interface for interacting with simapp",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "network chain ID")

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCmd(),
		txCmd(),
		flags.LineBreak,
		flags.LineBreak,
		keys.Commands(),
		flags.LineBreak,
		flags.NewCompletionCmd(rootCmd, true),
	)

	// Add flags and prefix all env exposed with GA
	executor := cli.PrepareMainCmd(rootCmd, "GA", simapp.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		fmt.Printf("Failed executing CLI command: %s, exiting...\n", err)
		os.Exit(1)
	}
}

func queryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(encodingConfig.Amino),
		flags.LineBreak,
		rpc.ValidatorCommand(encodingConfig.Amino),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(encodingConfig.Amino),
		authcmd.QueryTxCmd(encodingConfig.Amino),
		flags.LineBreak,
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
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		bankcmd.NewSendTxCmd(initClientCtx),
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
