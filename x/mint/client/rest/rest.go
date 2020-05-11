package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// RegisterRoutes registers minting module REST handlers on the provided router.
func RegisterRoutes(cliCtx context.Context, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
}
