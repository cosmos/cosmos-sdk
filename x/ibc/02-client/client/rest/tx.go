package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/client/client", createClientHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/client/clients/{%s}/update", RestClientID), updateClientHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/ibc/client/clients/{%s}/misbehaviour", RestClientID), submitMisbehaviourHandlerFn(cliCtx)).Methods("POST")
}

func createClientHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func updateClientHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func submitMisbehaviourHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
