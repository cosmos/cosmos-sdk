// Package grpcgateway utilizes the global gogoproto registry to create dynamic query handlers on net/http's mux router.
//
// Header `x-cosmos-block-height` allows you to specify a height for the query.
//
// Requests that do not have a dynamic handler registered will be routed to the canonical gRPC-Gateway mux.
package grpcgateway
