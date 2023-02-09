package module_test

import (
	"cosmossdk.io/core/appmodule"
)

// CoreAppModule is solely here for the purpose of generating
// mocks to be used in module tests.
type CoreAppModule interface {
	appmodule.AppModule
	appmodule.HasGenesis
	appmodule.HasBeginBlocker
	appmodule.HasEndBlocker
}
