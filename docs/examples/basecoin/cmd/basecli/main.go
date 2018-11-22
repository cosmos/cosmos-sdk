package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/docs/examples/basecoin/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	stake "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	storeAcc      = "acc"
	storeSlashing = "slashing"
	storeStake    = "stake"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "basecli",
		Short: "Basecoin light-client",
	}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// Setup certain SDK config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("baseacc", "basepub")
	config.SetBech32PrefixForValidator("baseval", "basevalpub")
	config.SetBech32PrefixForConsensusNode("basecons", "baseconspub")
	config.Seal()

	rs := lcd.NewRestServer(cdc)

	registerRoutes(rs)

	// TODO: Setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc.

	// add standard rpc, and tx commands
	rootCmd.AddCommand(
		rpc.InitClientCommand(),
		rpc.StatusCommand(),
		client.LineBreak,
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
	)

	// add query/post commands (custom to binary)
	rootCmd.AddCommand(
		stakecmd.GetCmdQueryValidator(storeStake, cdc),
		stakecmd.GetCmdQueryValidators(storeStake, cdc),
		stakecmd.GetCmdQueryValidatorUnbondingDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryValidatorRedelegations(storeStake, cdc),
		stakecmd.GetCmdQueryDelegation(storeStake, cdc),
		stakecmd.GetCmdQueryDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryPool(storeStake, cdc),
		stakecmd.GetCmdQueryParams(storeStake, cdc),
		stakecmd.GetCmdQueryUnbondingDelegation(storeStake, cdc),
		stakecmd.GetCmdQueryUnbondingDelegations(storeStake, cdc),
		stakecmd.GetCmdQueryRedelegation(storeStake, cdc),
		stakecmd.GetCmdQueryRedelegations(storeStake, cdc),
		slashingcmd.GetCmdQuerySigningInfo(storeSlashing, cdc),
		authcmd.GetAccountCmd(storeAcc, cdc),
	)

	rootCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		ibccmd.IBCTransferCmd(cdc),
		ibccmd.IBCRelayCmd(cdc),
		stakecmd.GetCmdCreateValidator(cdc),
		stakecmd.GetCmdEditValidator(cdc),
		stakecmd.GetCmdDelegate(cdc),
		stakecmd.GetCmdUnbond(storeStake, cdc),
		stakecmd.GetCmdRedelegate(storeStake, cdc),
		slashingcmd.GetCmdUnjail(cdc),
	)

	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		rs.ServeCommand(),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "BC", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		// Note: Handle with #870
		panic(err)
	}
}

func registerRoutes(rs *lcd.RestServer) {
	keys.RegisterRoutes(rs.Mux, rs.CliCtx.Indent)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, storeAcc)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	stake.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashing.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
}
