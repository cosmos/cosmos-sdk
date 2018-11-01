package main

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	distrcmd "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	govcmd "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"

	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
)

const (
	storeAcc        = "acc"
	storeGov        = "gov"
	storeSlashing   = "slashing"
	storeStake      = "stake"
	queryRouteStake = "stake"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "gaiacli",
		Short: "Gaia light-client",
	}
)

func main() {
	cobra.EnableCommandSorting = false
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc
	rootCmd.AddCommand(client.ConfigCmd())

	// add standard rpc commands
	rpc.AddCommands(rootCmd)

	//Add query commands
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}
	queryCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(queryCmd, cdc)
	queryCmd.AddCommand(client.LineBreak)
	queryCmd.AddCommand(client.GetCommands(
		authcmd.GetAccountCmd(storeAcc, cdc, authcmd.GetAccountDecoder(cdc)),
		stakecmd.GetCmdQueryDelegation(storeStake, cdc),
		stakecmd.GetCmdQueryDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryUnbondingDelegation(storeStake, cdc),
		stakecmd.GetCmdQueryUnbondingDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryRedelegation(storeStake, cdc),
		stakecmd.GetCmdQueryRedelegations(storeStake, cdc),
		stakecmd.GetCmdQueryValidator(storeStake, cdc),
		stakecmd.GetCmdQueryValidators(storeStake, cdc),
		stakecmd.GetCmdQueryValidatorUnbondingDelegations(queryRouteStake, cdc),
		stakecmd.GetCmdQueryValidatorRedelegations(queryRouteStake, cdc),
		stakecmd.GetCmdQueryParams(storeStake, cdc),
		stakecmd.GetCmdQueryPool(storeStake, cdc),
		govcmd.GetCmdQueryProposal(storeGov, cdc),
		govcmd.GetCmdQueryProposals(storeGov, cdc),
		govcmd.GetCmdQueryVote(storeGov, cdc),
		govcmd.GetCmdQueryVotes(storeGov, cdc),
		govcmd.GetCmdQueryDeposit(storeGov, cdc),
		govcmd.GetCmdQueryDeposits(storeGov, cdc),
		slashingcmd.GetCmdQuerySigningInfo(storeSlashing, cdc),
	)...)

	//Add query commands
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	//Add auth and bank commands
	txCmd.AddCommand(
		client.PostCommands(
			bankcmd.GetBroadcastCommand(cdc),
			authcmd.GetSignCommand(cdc, authcmd.GetAccountDecoder(cdc)),
		)...)
	txCmd.AddCommand(client.LineBreak)

	txCmd.AddCommand(
		client.PostCommands(
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdRedelegate(storeStake, cdc),
			stakecmd.GetCmdUnbond(storeStake, cdc),
			distrcmd.GetCmdWithdrawRewards(cdc),
			distrcmd.GetCmdSetWithdrawAddr(cdc),
			govcmd.GetCmdDeposit(cdc),
			bankcmd.SendTxCmd(cdc),
			govcmd.GetCmdSubmitProposal(cdc),
			slashingcmd.GetCmdUnjail(cdc),
			govcmd.GetCmdVote(cdc),
		)...)
	rootCmd.AddCommand(
		queryCmd,
		txCmd,
		lcd.ServeCommand(cdc),
		client.LineBreak,
	)

	// add proxy, version and key info
	rootCmd.AddCommand(
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "GA", app.DefaultCLIHome)
	err := initConfig(rootCmd)
	if err != nil {
		panic(err)
	}

	err = executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
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

	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
