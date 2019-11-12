package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/connection/connections/{%s}", RestConnectionID), queryConnectionHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/connection/clients/{%s}/connections", RestClientID), queryClientConnectionsHandlerFn(cliCtx)).Methods("GET")
}

func queryConnectionHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func queryClientConnectionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
