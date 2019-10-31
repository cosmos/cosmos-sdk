package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/gorilla/mux"
)

// REST query and parameter values
const (
	RestParamEvidenceHash = "evidence-hash"

	MethodGet = "GET"
)

// EvidenceRESTHandler defines a REST service evidence handler implemented in
// another module. The sub-route is mounted on the evidence REST handler.
type EvidenceRESTHandler struct {
	SubRoute string
	Handler  func(http.ResponseWriter, *http.Request)
}

// RegisterRoutes registers all Evidence submission handlers for the evidence module's
// REST service handler.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, handlers []EvidenceRESTHandler) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r, handlers)
}
