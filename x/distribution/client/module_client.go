package client

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/utils"
	dist "github.com/cosmos/cosmos-sdk/x/distribution"
	distCmds "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	distQueryCmd := &cobra.Command{
		Use:                        dist.ModuleName,
		Short:                      "Querying commands for the distribution module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       utils.ValidateCmd,
	}

	distQueryCmd.AddCommand(client.GetCommands(
		distCmds.GetCmdQueryParams(mc.storeKey, mc.cdc),
		distCmds.GetCmdQueryValidatorOutstandingRewards(mc.storeKey, mc.cdc),
		distCmds.GetCmdQueryValidatorCommission(mc.storeKey, mc.cdc),
		distCmds.GetCmdQueryValidatorSlashes(mc.storeKey, mc.cdc),
		distCmds.GetCmdQueryDelegatorRewards(mc.storeKey, mc.cdc),
		distCmds.GetCmdQueryCommunityPool(mc.storeKey, mc.cdc),
	)...)

	return distQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:                        "distr",
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       utils.ValidateCmd,
	}

	distTxCmd.AddCommand(client.PostCommands(
		distCmds.GetCmdWithdrawRewards(mc.cdc),
		distCmds.GetCmdSetWithdrawAddr(mc.cdc),
		distCmds.GetCmdWithdrawAllRewards(mc.cdc, mc.storeKey),
	)...)

	return distTxCmd
}
