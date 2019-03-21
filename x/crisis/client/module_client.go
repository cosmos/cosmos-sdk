package client

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/crisis/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
	routes   crisis.Routes
}

// NewModuleClient creates a new ModuleClient object
func NewModuleClient(storeKey string, cdc *amino.Codec, routes crisis.Routes) ModuleClient {
	return ModuleClient{
		storeKey: storeKey,
		cdc:      cdc,
		routes:   routes,
	}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	slashingQueryCmd := &cobra.Command{
		Use:   crisis.ModuleName,
		Short: "Querying commands for the crisis module",
	}

	slashingQueryCmd.AddCommand(
		client.GetCommands(
			cli.GetCmdQuerySigningInfo(mc.storeKey, mc.cdc, mc.routes),
		)...,
	)

	return slashingQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   crisis.ModuleName,
		Short: "crisis transactions subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		cli.GetCmdInvariantBroken(mc.cdc),
	)...)
	return txCmd
}
