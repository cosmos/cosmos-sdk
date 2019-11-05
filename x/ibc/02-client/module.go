package client

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/client/cli"
)

// Name returns the IBC client name
func Name() string {
	return SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC client
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	// TODO:
}

// GetTxCmd returns the root tx command for the IBC client
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	return cli.GetTxCmd(fmt.Sprintf("%s/%s", storeKey, SubModuleName), cdc)
}

// GetQueryCmd returns no root query command for the IBC client
func GetQueryCmd(cdc *codec.Codec, queryRoute string) *cobra.Command {
	return cli.GetQueryCmd(fmt.Sprintf("%s/%s", queryRoute, SubModuleName), cdc)
}
