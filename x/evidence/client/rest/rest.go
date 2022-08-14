package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rest"

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
func RegisterRoutes(clientCtx client.Context, rtr *mux.Router, handlers []EvidenceRESTHandler) {
	r := rest.WithHTTPDeprecationHeaders(rtr)

	registerQueryRoutes(clientCtx, r)
	registerTxRoutes(clientCtx, r, handlers)
}
