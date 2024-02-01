package runtime

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
)

// NewEnvironment creates a new environment for the application
// if memstoreservice is needed, it can be added to the environment: environment.MemStoreService = memstoreservice
func NewEnvironment(kvService store.KVStoreService) appmodule.Environment {
	return appmodule.Environment{
		EventService:   EventService{},
		HeaderService:  HeaderService{},
		BranchService:  BranchService{},
		GasService:     GasService{},
		KVStoreService: kvService,
	}
}
