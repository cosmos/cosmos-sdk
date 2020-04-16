package tendermint

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/client/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC client
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router, queryRoute string) {
	rest.RegisterRoutes(ctx, rtr, fmt.Sprintf("%s/%s", queryRoute, types.SubModuleName))
}

// GetTxCmd returns the root tx command for the IBC client
func GetTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	return cli.NewTxCmd(m, txg, ar)
}
