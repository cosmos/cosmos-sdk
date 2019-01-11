package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/docs/examples/basecoin/app"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	at "github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	sl "github.com/cosmos/cosmos-sdk/x/slashing"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	st "github.com/cosmos/cosmos-sdk/x/staking"
	stakingcmd "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	staking "github.com/cosmos/cosmos-sdk/x/staking/client/rest"
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

	// TODO: Setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc.

	// add standard rpc, and tx commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		client.LineBreak,
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
	)

	// add query/post commands (custom to binary)
	rootCmd.AddCommand(
		stakingcmd.GetCmdQueryValidator(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryValidators(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryValidatorUnbondingDelegations(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryValidatorRedelegations(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryDelegation(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryDelegations(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryPool(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryParams(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryUnbondingDelegation(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryUnbondingDelegations(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryRedelegation(st.StoreKey, cdc),
		stakingcmd.GetCmdQueryRedelegations(st.StoreKey, cdc),
		slashingcmd.GetCmdQuerySigningInfo(sl.StoreKey, cdc),
		stakingcmd.GetCmdQueryValidatorDelegations(st.StoreKey, cdc),
		authcmd.GetAccountCmd(at.StoreKey, cdc),
	)

	rootCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		ibccmd.IBCTransferCmd(cdc),
		ibccmd.IBCRelayCmd(cdc),
		stakingcmd.GetCmdCreateValidator(cdc),
		stakingcmd.GetCmdEditValidator(cdc),
		stakingcmd.GetCmdDelegate(cdc),
		stakingcmd.GetCmdUnbond(st.StoreKey, cdc),
		stakingcmd.GetCmdRedelegate(st.StoreKey, cdc),
		slashingcmd.GetCmdUnjail(cdc),
	)

	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc, registerRoutes),
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
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, at.StoreKey)
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	staking.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashing.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
}
