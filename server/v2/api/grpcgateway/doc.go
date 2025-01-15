// Package grpcgateway provides a custom http mux that utilizes the global gogoproto registry to match
// to create dynamic query handlers.
//
// Header `x-cosmos-block-height` allows you to specify a height for the query.
//
// Requests that do not have a dynamic handler will be routed to the canonical gRPC gateway mux.
package grpcgateway
