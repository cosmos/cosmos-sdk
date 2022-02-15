package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

// DeprecationURL is the URL for migrating deprecated REST endpoints to newer ones.
// TODO Switch to `/` (not `/master`) once v0.40 docs are deployed.
// https://github.com/cosmos/cosmos-sdk/issues/8019
const DeprecationURL = "https://docs.cosmos.network/master/migrations/rest.html"

// addHTTPDeprecationHeaders is a mux middleware function for adding HTTP
// Deprecation headers to a http handler
func addHTTPDeprecationHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", "true")
		w.Header().Set("Link", "<"+DeprecationURL+">; rel=\"deprecation\"")
		w.Header().Set("Warning", "199 - \"this endpoint is deprecated and may not work as before, see deprecation link for more info\"")
		h.ServeHTTP(w, r)
	})
}

// nolint
// WithHTTPDeprecationHeaders returns a new *mux.Router, identical to its input
// but with the addition of HTTP Deprecation headers. This is used to mark legacy
// amino REST endpoints as deprecated in the REST API.
// nolint: gocritic
func WithHTTPDeprecationHeaders(r *mux.Router) *mux.Router {
	subRouter := r.NewRoute().Subrouter()
	subRouter.Use(addHTTPDeprecationHeaders)
	return subRouter
}
