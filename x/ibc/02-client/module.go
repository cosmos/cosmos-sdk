package ics02

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/cli"
)

// Name returns the staking module's name
func Name() string {
	return SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the staking module.
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	// TODO:
	// rest.RegisterRoutes(ctx, rtr)
}

// GetTxCmd returns the root tx command for the staking module.
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	return cli.GetTxCmd(fmt.Sprintf("%s/%s", storeKey, SubModuleName), cdc)
}

// GetQueryCmd returns no root query command for the staking module.
func GetQueryCmd(cdc *codec.Codec, queryRoute string) *cobra.Command {
	return cli.GetQueryCmd(fmt.Sprintf("%s/%s", queryRoute, SubModuleName), cdc)
}
