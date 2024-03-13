// module_test inconsistenctly import appmodulev2 & appmodulev1 due to limitation in mockgen
// eventually, when we change mocking library, we should be consistent in our appmodule imports
package module_test

import (
	"cosmossdk.io/core/appmodule"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// AppModuleWithAllExtensions is solely here for the purpose of generating
// mocks to be used in module tests.
type AppModuleWithAllExtensions interface {
	module.AppModule
	module.HasServices
	appmodulev2.HasGenesis
	module.HasInvariants
	module.HasConsensusVersion
	module.HasABCIEndBlock
	module.HasName
}

// mocks to be used in module tests.
type AppModuleWithAllExtensionsABCI interface {
	module.AppModule
	module.HasServices
	module.HasABCIGenesis
	module.HasInvariants
	module.HasConsensusVersion
	module.HasABCIEndBlock
	module.HasName
}

// CoreAppModule is solely here for the purpose of generating
// mocks to be used in module tests.
type CoreAppModule interface {
	appmodulev2.AppModule
	appmodule.HasGenesisAuto
	appmodulev2.HasBeginBlocker
	appmodulev2.HasEndBlocker
	appmodule.HasPrecommit
	appmodule.HasPrepareCheckState
}

type CoreAppModuleWithPreBlock interface {
	CoreAppModule
	appmodulev2.HasPreBlocker
}
