package connection

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/rest"
)

// Name returns the IBC connection ICS name
func Name() string {
	return SubModuleName
}

// GetTxCmd returns the root tx command for the IBC connections.
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	return cli.GetTxCmd(fmt.Sprintf("%s/%s", storeKey, SubModuleName), cdc)
}

// GetQueryCmd returns no root query command for the IBC connections.
func GetQueryCmd(cdc *codec.Codec, queryRoute string) *cobra.Command {
	return cli.GetQueryCmd(fmt.Sprintf("%s/%s", queryRoute, SubModuleName), cdc)
}

// RegisterRESTRoutes registers the REST routes for the IBC connections.
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router, queryRoute string) {
	rest.RegisterRoutes(ctx, rtr, fmt.Sprintf("%s/%s", queryRoute, SubModuleName))
}
