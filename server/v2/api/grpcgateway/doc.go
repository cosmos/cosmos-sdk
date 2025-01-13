// Package grpcgateway provides a custom http mux that utilizes the global gogoproto registry to match
// grpc gateway requests to query handlers. POST requests with JSON bodies and GET requests with query params are supported.
// Wildcard endpoints (i.e. foo/bar/{baz}), as well as catch-all endpoints (i.e. foo/bar/{baz=**} are supported. Using
// header `x-cosmos-block-height` allows you to specify a height for the query.
//
// The URL matching logic is achieved by building regular expressions from the gateway HTTP annotations. These regular expressions
// are then used to match against incoming requests to the HTTP server.
//
// In cases where the custom http mux is unable to handle the query (i.e. no match found), the request will fall back to the
// ServeMux from github.com/grpc-ecosystem/grpc-gateway/runtime.
package grpcgateway
