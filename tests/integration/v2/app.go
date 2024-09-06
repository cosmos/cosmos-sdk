package integration

import (
	"fmt"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"

	"github.com/cosmos/cosmos-sdk/codec"
)

func NewIntegrationApp(
	logger log.Logger,
	addressCodec address.Codec,
	validatorCodec address.ValidatorAddressCodec,
	modules map[string]appmodule.AppModule,
) {
	interfaceRegistry, _, err := codec.ProvideInterfaceRegistry(
		addressCodec,
		validatorCodec,
		nil,
	)
	if err != nil {
		panic(err)
	}

	legacyAmino := codec.ProvideLegacyAmino()
	appBuilder, routerBuilder, _, _, _ := runtime.ProvideAppBuilder[transaction.Tx](
		interfaceRegistry,
		legacyAmino,
	)
	mm := runtime.NewModuleManager[transaction.Tx](
		logger,
		&runtimev2.Module{},
		modules,
	)
	// appBuilder.RegisterModules(mm)
	fmt.Printf("appBuilder: %v\nrouterBuilder: %v\nmodule manager: %v",
		appBuilder,
		routerBuilder,
		mm)
	app, err := appBuilder.Build()
	if err != nil {
		panic(err)
	}
	fmt.Printf("appBuilder: %v\nrouterBuilder: %v\nmodule manager: %v\napp: %v",
		appBuilder,
		routerBuilder,
		mm,
		app)
}
