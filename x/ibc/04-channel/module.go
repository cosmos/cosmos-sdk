package channel

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/rest"
)

// Name returns the IBC connection ICS name
func Name() string {
	return SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC channel
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router, queryRoute string) {
	rest.RegisterRoutes(ctx, rtr, fmt.Sprintf("%s/%s", queryRoute, SubModuleName))
}

// GetTxCmd returns the root tx command for the IBC connections.
func GetTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	return cli.NewTxCmd(m, txg, ar)
}

// GetQueryCmd returns no root query command for the IBC connections.
func GetQueryCmd(cdc *codec.Codec, queryRoute string) *cobra.Command {
	return cli.GetQueryCmd(fmt.Sprintf("%s/%s", queryRoute, SubModuleName), cdc)
}
