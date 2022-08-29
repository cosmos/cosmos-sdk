package appmodule

import (
	"google.golang.org/grpc"

	"cosmossdk.io/core/blockinfo"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
)

// Service bundles all the core services provided by a compliant runtime
// implementation into a single services.
//
// NOTE: If new core services are provided, they shouldn't be added to this
// interface which would be API breaking, but rather a cosmossdk.io/core/appmodule/v2.Service
// should be created which extends this Service interface with new services.
type Service interface {
	store.KVStoreService
	store.MemoryStoreService
	store.TransientStoreService
	event.Service
	blockinfo.Service
	gas.Service
	InterModuleClient
}

// InterModuleClient is an inter-module client as specified in ADR-033. It
// allows one module to send msg's and queries to other modules provided
// that the request is valid and can be properly authenticated. This client
// by default allows sending messages authenticated with the ADR-028 root module
// key.
type InterModuleClient interface {
	grpc.ClientConnInterface

	// DerivedClient returns an inter-module client for the ADR-028 derived
	// module address for the provided key.
	DerivedClient(key []byte) grpc.ClientConnInterface
}
