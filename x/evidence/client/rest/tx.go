package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx context.Context, r *mux.Router, handlers []EvidenceRESTHandler) {
	// TODO: Register tx handlers.
}
