package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/connection/connection/open-init", connectionOpenInitHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/ibc/connection/connection/open-try", connectionOpenTryHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/connection/connections/{%s}/open-ack", RestClientID), connectionOpenAckHandlerFn(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/ibc/connection/connections/{%s}/open-confirm", RestClientID), connectionOpenConfirmHandlerFn(cliCtx)).Methods("PUT")
}

func connectionOpenInitHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func connectionOpenTryHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func connectionOpenAckHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func connectionOpenConfirmHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
