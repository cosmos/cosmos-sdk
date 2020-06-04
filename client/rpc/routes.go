package rpc

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// Register REST endpoints.
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/node_info", NodeInfoRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/syncing", NodeSyncingRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/blocks/latest", LatestBlockRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/blocks/{height}", BlockRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/validatorsets/latest", LatestValidatorSetRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/validatorsets/{height}", ValidatorSetRequestHandlerFn(clientCtx)).Methods("GET")
}
