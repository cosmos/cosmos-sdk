package client

import "net/http"

func AddDeprecationHeaders(rw http.ResponseWriter) http.ResponseWriter {
	rw.Header().Set("Deprecation", "true")
	rw.Header().Set("Link", "<https://docs.cosmos.network/v0.40/interfaces/rest.html>; rel=\"deprecation\"")
	return rw
}
