package module_test

import (
	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// AppModuleWithAllExtensions is solely here for the purpose of generating
// mocks to be used in module tests.
type AppModuleWithAllExtensions interface {
	module.AppModule
	module.HasServices
	module.HasGenesis
	module.HasInvariants
	module.HasConsensusVersion
	module.HasABCIEndblock
}

// CoreAppModule is solely here for the purpose of generating
// mocks to be used in module tests.
type CoreAppModule interface {
	appmodule.AppModule
	appmodule.HasGenesis
	appmodule.HasBeginBlocker
	appmodule.HasEndBlocker
	appmodule.HasPrecommit
	appmodule.HasPrepareCheckState
}
