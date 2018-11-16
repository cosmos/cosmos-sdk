package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	_ "github.com/cosmos/cosmos-sdk/client/lcd/statik"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/docs/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/docs/examples/basecoin/types"
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
		client.GetCommands(
			stakecmd.GetCmdQueryValidator("stake", cdc),
			stakecmd.GetCmdQueryValidators("stake", cdc),
			stakecmd.GetCmdQueryValidatorUnbondingDelegations("stake", cdc),
			stakecmd.GetCmdQueryValidatorRedelegations("stake", cdc),
			stakecmd.GetCmdQueryDelegation("stake", cdc),
			stakecmd.GetCmdQueryDelegations("stake", cdc),
			stakecmd.GetCmdQueryPool("stake", cdc),
			stakecmd.GetCmdQueryParams("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegation("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegations("stake", cdc),
			stakecmd.GetCmdQueryRedelegation("stake", cdc),
			stakecmd.GetCmdQueryRedelegations("stake", cdc),
			slashingcmd.GetCmdQuerySigningInfo("slashing", cdc),
			authcmd.GetAccountCmd("acc", cdc, types.GetAccountDecoder(cdc)),
		)...)

	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
			ibccmd.IBCTransferCmd(cdc),
			ibccmd.IBCRelayCmd(cdc),
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdUnbond("stake", cdc),
			stakecmd.GetCmdRedelegate("stake", cdc),
			slashingcmd.GetCmdUnjail(cdc),
		)...)

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
	auth.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, "acc")
	bank.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	stake.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashing.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
}
