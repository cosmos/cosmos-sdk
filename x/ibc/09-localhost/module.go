package localhost

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/client/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}

// RegisterRESTRoutes registers the REST routes for the IBC localhost client
func RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router, queryRoute string) {
	rest.RegisterRoutes(ctx, rtr, fmt.Sprintf("%s/%s", queryRoute, types.SubModuleName))
}

// GetTxCmd returns the root tx command for the IBC localhost client
func GetTxCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	return cli.GetTxCmd(cdc, fmt.Sprintf("%s/%s", storeKey, types.SubModuleName))
}
