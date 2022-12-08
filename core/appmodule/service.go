package appmodule

import (
	"cosmossdk.io/core/intermodule"
	"cosmossdk.io/core/store"
)

// Service bundles all the core services provided by a compliant runtime
// implementation into a single services.
//
// NOTE: If new core services are provided after core v.10, they shouldn't be added to this
// interface which would be API breaking, but rather a cosmossdk.io/core/appmodule/v2.Service
// should be created which extends this Service interface with new services.
type Service interface {
	store.KVStoreService
	store.MemoryStoreService
	store.TransientStoreService
	intermodule.Client
}
