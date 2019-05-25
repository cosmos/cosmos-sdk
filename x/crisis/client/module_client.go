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
}

// NewModuleClient creates a new ModuleClient object
func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// GetQueryCmd returns the cli query commands for this module
func (ModuleClient) GetQueryCmd() *cobra.Command {
	return nil
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
