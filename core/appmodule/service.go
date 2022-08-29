package appmodule

import (
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
