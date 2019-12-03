package transfer

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/client/rest"
)

// Name returns the IBC transfer ICS name
func Name() string {
	return SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC transfer
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

// GetTxCmd returns the root tx command for the IBC transfer.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

// GetQueryCmd returns the root tx command for the IBC transfer.
func GetQueryCmd(cdc *codec.Codec, queryRoute string) *cobra.Command {
	return cli.GetQueryCmd(cdc, queryRoute)
}
