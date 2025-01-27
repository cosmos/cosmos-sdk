package appconfig

import (
	"fmt"
	"reflect"

	"github.com/cosmos/gogoproto/proto"

	internal "cosmossdk.io/depinject/internal/appconfig"
)

var Register = RegisterModule

// RegisterModule registers a module with the global module registry. The provided
// protobuf message is used only to uniquely identify the protobuf module config
// type. The instance of the protobuf message used in the actual configuration
// will be injected into the container and can be requested by a provider
// function. All module initialization should be handled by the provided options.
//
// Config is a protobuf message type. It should define the cosmos.app.v1alpha.module
// option and must explicitly specify go_packageto make debugging easier for users.
//
// If you want to customize an existing module, you need to overwrite by calling
// RegisterModule again with the same config (proto API type) and new Provide or
// Invoke options. Example:
//
// - Create a new struct and wrap the existing module inside it:
//
//	type MyBankAppModule struct {
//	  bank.AppModule // core bank module
//	  // additional helper fields
//	  cdc codec.Codec
//	}
//
// - Overwrite function that you want to customize (eg DefaultGenesis).
// - Create a new Provide function (to provide new Module implementation for the proto module):
//
//	func ProvideBankModule(in bank.ModuleInputs) BankModuleOutputs {
//	  // original provider that initializes the bank keeper
//	  pm := bank.ProvideModule(in)
//	  m := NewMyBankAppModule(in.Cdc, pm.BankKeeper, in.AccountKeeper)
//	  return BankModuleOutputs{
//	    BankKeeper: pm.BankKeeper,
//	    Module:     m,
//	  }
//	}
//
// - Re-register the bank module by using the original proto api module, and the new provider:
//
//	appconfig.RegisterModule(
//	  &bankmodulev1.Module{}, // from cosmossdk.io/api/cosmos/bank/module/v1
//	  appconfig.Provide(ProvideBankModule),
//	  appconfig.Invoke(bank.InvokeSetSendRestrictions),
//	)
func RegisterModule(config any, options ...Option) {
	protoConfig, ok := config.(proto.Message)
	if !ok {
		panic(fmt.Errorf("expected config to be a proto.Message, got %T", config))
	}

	ty := reflect.TypeOf(config)
	init := &internal.ModuleInitializer{
		ConfigProtoMessage: protoConfig,
		ConfigGoType:       ty,
	}
	internal.ModuleRegistry[ty] = init

	for _, option := range options {
		init.Error = option.apply(init)
		if init.Error != nil {
			return
		}
	}
}

// Option is a functional option for implementing modules.
type Option interface {
	apply(*internal.ModuleInitializer) error
}

type funcOption func(initializer *internal.ModuleInitializer) error

func (f funcOption) apply(initializer *internal.ModuleInitializer) error {
	return f(initializer)
}

// Provide registers providers with the dependency injection system that will be
// run within the module scope (depinject.ProvideInModule). See cosmossdk.io/depinject for
// documentation on the dependency injection system.
func Provide(providers ...interface{}) Option {
	return funcOption(func(initializer *internal.ModuleInitializer) error {
		initializer.Providers = append(initializer.Providers, providers...)
		return nil
	})
}

// Invoke registers invokers to run with depinject (depinject.InvokeInModule). Each invoker will be called
// at the end of dependency graph configuration in the order in which it was defined. Invokers may not define output
// parameters, although they may return an error, and all of their input parameters will be marked as optional so that
// invokers impose no additional constraints on the dependency graph. Invoker functions should nil-check all inputs.
func Invoke(invokers ...interface{}) Option {
	return funcOption(func(initializer *internal.ModuleInitializer) error {
		initializer.Invokers = append(initializer.Invokers, invokers...)
		return nil
	})
}
