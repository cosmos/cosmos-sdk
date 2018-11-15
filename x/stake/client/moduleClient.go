package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct{}

func NewModuleClient() ModuleClient {
	return ModuleClient{}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	stakeQueryCmd := &cobra.Command{
		Use:   "stake",
		Short: "Querying commands for the staking module",
	}
	stakeQueryCmd.AddCommand(client.GetCommands(
		cli.GetCmdQueryDelegation(storeKey, cdc),
		cli.GetCmdQueryDelegations(storeKey, cdc),
		cli.GetCmdQueryUnbondingDelegation(storeKey, cdc),
		cli.GetCmdQueryUnbondingDelegations(storeKey, cdc),
		cli.GetCmdQueryRedelegation(storeKey, cdc),
		cli.GetCmdQueryRedelegations(storeKey, cdc),
		cli.GetCmdQueryValidator(storeKey, cdc),
		cli.GetCmdQueryValidators(storeKey, cdc),
		cli.GetCmdQueryValidatorDelegations(storeKey, cdc),
		cli.GetCmdQueryValidatorUnbondingDelegations(storeKey, cdc),
		cli.GetCmdQueryValidatorRedelegations(storeKey, cdc),
		cli.GetCmdQueryParams(storeKey, cdc),
		cli.GetCmdQueryPool(storeKey, cdc))...)

	return stakeQueryCmd

}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	stakeTxCmd := &cobra.Command{
		Use:   "stake",
		Short: "Staking transaction subcommands",
	}

	stakeTxCmd.AddCommand(client.PostCommands(
		cli.GetCmdCreateValidator(cdc),
		cli.GetCmdEditValidator(cdc),
		cli.GetCmdDelegate(cdc),
		cli.GetCmdRedelegate(storeKey, cdc),
		cli.GetCmdUnbond(storeKey, cdc),
	)...)

	return stakeTxCmd
}
