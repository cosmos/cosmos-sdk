package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	govcmd "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"path"
	"os"
	"github.com/spf13/viper"
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

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc
	rootCmd.AddCommand(client.ConfigCmd())

	// add standard rpc commands
	rpc.AddCommands(rootCmd)

	//Add state commands
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint state querying subcommands",
	}
	tendermintCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(tendermintCmd, cdc)

	rootCmd.AddCommand(
		tendermintCmd,
		lcd.ServeCommand(cdc),
		client.LineBreak,
	)

	//Add stake commands
	stakeCmd := &cobra.Command{
		Use:   "stake",
		Short: "Stake and validation subcommands",
	}
	stakeCmd.AddCommand(
		client.GetCommands(
			stakecmd.GetCmdQueryValidator("stake", cdc),
			stakecmd.GetCmdQueryValidators("stake", cdc),
			stakecmd.GetCmdQueryDelegation("stake", cdc),
			stakecmd.GetCmdQueryDelegations("stake", cdc),
			stakecmd.GetCmdQueryParams("stake", cdc),
			stakecmd.GetCmdQueryPool("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegation("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegations("stake", cdc),
			stakecmd.GetCmdQueryRedelegation("stake", cdc),
			stakecmd.GetCmdQueryRedelegations("stake", cdc),
			slashingcmd.GetCmdQuerySigningInfo("slashing", cdc),
		)...)
	stakeCmd.AddCommand(
		client.PostCommands(
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdUnbond("stake", cdc),
			stakecmd.GetCmdRedelegate("stake", cdc),
			slashingcmd.GetCmdUnjail(cdc),
		)...)
	rootCmd.AddCommand(
		stakeCmd,
	)

	//Add stake commands
	govCmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance and voting subcommands",
	}
	govCmd.AddCommand(
		client.GetCommands(
			govcmd.GetCmdQueryProposal("gov", cdc),
			govcmd.GetCmdQueryVote("gov", cdc),
			govcmd.GetCmdQueryVotes("gov", cdc),
			govcmd.GetCmdQueryProposals("gov", cdc),
		)...)
	govCmd.AddCommand(
		client.PostCommands(
			govcmd.GetCmdSubmitProposal(cdc),
			govcmd.GetCmdDeposit(cdc),
			govcmd.GetCmdVote(cdc),
		)...)
	rootCmd.AddCommand(
		govCmd,
	)

	//Add auth and bank commands
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
			authcmd.GetSignCommand(cdc, authcmd.GetAccountDecoder(cdc)),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
			bankcmd.GetBroadcastCommand(cdc),
		)...)

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
