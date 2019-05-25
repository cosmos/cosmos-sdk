package rest

import (
	"github.com/YunSuk-Yeo/cosmos-sdk/client/context"
	"github.com/YunSuk-Yeo/cosmos-sdk/codec"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers minting module REST handlers on the provided router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	registerQueryRoutes(cliCtx, r, cdc)
}
