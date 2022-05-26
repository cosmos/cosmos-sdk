package runtime

import "github.com/cosmos/cosmos-sdk/types/module"

// AppModuleWrapper is a type used for injecting a module.AppModule into a
// container so that it can be used by the runtime module.
type AppModuleWrapper struct {
	module.AppModule
}

// WrapAppModule wraps a module.AppModule so that it can be injected into
// a container for use by the runtime module.
func WrapAppModule(appModule module.AppModule) AppModuleWrapper {
	return AppModuleWrapper{AppModule: appModule}
}

// IsOnePerModuleType identifies this type as a depinject.OnePerModuleType.
func (AppModuleWrapper) IsOnePerModuleType() {}

// AppModuleBasicWrapper is a type used for injecting a module.AppModuleBasic
// into a container so that it can be used by the runtime module.
type AppModuleBasicWrapper struct {
	module.AppModuleBasic
}

// WrapAppModuleBasic wraps a module.AppModuleBasic so that it can be injected into
// a container for use by the runtime module.
func WrapAppModuleBasic(basic module.AppModuleBasic) AppModuleBasicWrapper {
	return AppModuleBasicWrapper{AppModuleBasic: basic}
}

// IsOnePerModuleType identifies this type as a depinject.OnePerModuleType.
func (AppModuleBasicWrapper) IsOnePerModuleType() {}
