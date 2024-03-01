package runtime

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// NewEnvironment creates a new environment for the application
// For setting custom services that aren't set by default, use the EnvOption
// Note: Depinject always provide an environment with all services (mandatory and optional)
func NewEnvironment(
	kvService store.KVStoreService,
	logger log.Logger,
	opts ...EnvOption,
) appmodule.Environment {
	env := appmodule.Environment{
		Logger:         logger,
		EventService:   EventService{},
		HeaderService:  HeaderService{},
		BranchService:  BranchService{},
		GasService:     GasService{},
		KVStoreService: kvService,
	}

	for _, opt := range opts {
		opt(&env)
	}

	return env
}

type EnvOption func(*appmodule.Environment)

func EnvWithRouterService(queryServiceRouter *baseapp.GRPCQueryRouter, msgServiceRouter *baseapp.MsgServiceRouter) EnvOption {
	return func(env *appmodule.Environment) {
		env.RouterService = NewRouterService(env.KVStoreService, queryServiceRouter, msgServiceRouter)
	}
}

func EnvWithMemStoreService(memStoreService store.MemoryStoreService) EnvOption {
	return func(env *appmodule.Environment) {
		env.MemStoreService = memStoreService
	}
}
