package rest

import (
	"net/http"
)

// EvidenceRESTHandler defines a REST service evidence handler implemented in
// another module. The sub-route is mounted on the evidence REST handler.
type EvidenceRESTHandler struct {
	SubRoute string
	Handler  func(http.ResponseWriter, *http.Request)
}
