package runtime

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
)

// NewEnvironment creates a new environment for the application
// if memstoreservice is needed, it can be added to the environment: environment.MemStoreService = memstoreservice
func NewEnvironment(kvService store.KVStoreService, logger log.Logger) appmodule.Environment {
	return appmodule.Environment{
		EventService:   EventService{},
		HeaderService:  HeaderService{},
		BranchService:  BranchService{},
		GasService:     GasService{},
		KVStoreService: kvService,
		Logger:         logger,
	}
}
