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

	//Add query commands
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Querying subcommands",
	}
	queryCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(queryCmd, cdc)
	queryCmd.AddCommand(client.LineBreak)
	queryCmd.AddCommand(client.GetCommands(
		authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
		stakecmd.GetCmdQueryDelegation("stake", cdc),
		stakecmd.GetCmdQueryDelegations("stake", cdc),
		stakecmd.GetCmdQueryParams("stake", cdc),
		stakecmd.GetCmdQueryPool("stake", cdc),
		govcmd.GetCmdQueryProposal("gov", cdc),
		govcmd.GetCmdQueryProposals("gov", cdc),
		stakecmd.GetCmdQueryRedelegation("stake", cdc),
		stakecmd.GetCmdQueryRedelegations("stake", cdc),
		slashingcmd.GetCmdQuerySigningInfo("slashing", cdc),
		stakecmd.GetCmdQueryUnbondingDelegation("stake", cdc),
		stakecmd.GetCmdQueryUnbondingDelegations("stake", cdc),
		stakecmd.GetCmdQueryValidator("stake", cdc),
		stakecmd.GetCmdQueryValidators("stake", cdc),
		govcmd.GetCmdQueryVote("gov", cdc),
		govcmd.GetCmdQueryVotes("gov", cdc),
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
		)...)
	txCmd.AddCommand(
		authcmd.GetSignCommand(cdc, authcmd.GetAccountDecoder(cdc)),
	)
	txCmd.AddCommand(client.LineBreak)

	txCmd.AddCommand(
		client.PostCommands(
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			govcmd.GetCmdDeposit(cdc),
			stakecmd.GetCmdRedelegate("stake", cdc),
			bankcmd.SendTxCmd(cdc),
			govcmd.GetCmdSubmitProposal(cdc),
			stakecmd.GetCmdUnbond("stake", cdc),
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
