package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

<<<<<<< HEAD:x/ibc/02-client/client/rest/rest.go
// REST client flags
const (
	RestClientID   = "client-id"
	RestRootEpoch  = "epoch"
	RestRootHeight = "height"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
=======
// RegisterRoutes registers REST routes for the upgrade module under the path specified by routeName.
>>>>>>> d9fd4d2ca9a3f70fbabcd3eb6a1427395fdedf74:x/upgrade/client/rest/rest.go
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)
}
