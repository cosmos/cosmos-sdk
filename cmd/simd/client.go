package main

import (
	"os"
	"path"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

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
)

func init() {
	authclient.Codec = encodingConfig.Marshaler
}

func addClientCommands(rootClientCmd *cobra.Command) {

	// Add --chain-id to persistent flags and mark it required
	rootClientCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootClientCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootClientCmd)
	}

	// Construct Root Command
	rootClientCmd.AddCommand(
		rpc.StatusCommand(),
		client.ConfigCmd(simapp.DefaultNodeHome),
		queryCmd(encodingConfig),
		txCmd(encodingConfig),
		keys.Commands(),
		flags.NewCompletionCmd(rootClientCmd, true),
	)
}

func queryCmd(config simappparams.EncodingConfig) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cdc := config.Amino

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
	)

	// add modules' query commands
	clientCtx := client.Context{}
	clientCtx = clientCtx.
		WithJSONMarshaler(config.Marshaler).
		WithCodec(cdc)

	simapp.ModuleBasics.AddQueryCommands(queryCmd, clientCtx)

	return queryCmd
}

func txCmd(config simappparams.EncodingConfig) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cdc := config.Amino
	clientCtx := client.Context{}
	clientCtx = clientCtx.
		WithJSONMarshaler(config.Marshaler).
		WithTxGenerator(config.TxGenerator).
		WithAccountRetriever(types.NewAccountRetriever(config.Marshaler)).
		WithCodec(cdc)

	txCmd.AddCommand(
		bankcmd.NewSendTxCmd(clientCtx),
		flags.LineBreak,
		authcmd.GetSignCommand(clientCtx),
		authcmd.GetSignBatchCommand(cdc),
		authcmd.GetMultiSignCommand(clientCtx),
		authcmd.GetValidateSignaturesCommand(clientCtx),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(clientCtx),
		authcmd.GetEncodeCommand(clientCtx),
		authcmd.GetDecodeCommand(clientCtx),
		flags.LineBreak,
	)

	// add modules' tx commands
	simapp.ModuleBasics.AddTxCommands(txCmd, clientCtx)

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
