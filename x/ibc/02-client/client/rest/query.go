package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/client/consensus-state/clients/{%s}", RestClientID), queryConsensusStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/client/header", queryHeaderHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/client/state/clients/{%s}", RestClientID), queryClientStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/client/root/clients/{%s}/heights/{%s}", RestClientID, RestRootHeight), queryRootHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/client/node-state", queryNodeConsensusStateHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/ibc/client/path", queryPathHandlerFn(cliCtx)).Methods("GET")
}

func queryConsensusStateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func queryHeaderHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func queryClientStateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func queryRootHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func queryNodeConsensusStateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func queryPathHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
