package main

import (
	"os"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	rosettaCmd "cosmossdk.io/tools/rosetta/cmd"
	"cosmossdk.io/tools/rosetta/lib/logger"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/runtime"
)

var Config = appconfig.Compose(&appv1alpha1.Config{
	Modules: []*appv1alpha1.ModuleConfig{
		{
			Name: "runtime",
			Config: appconfig.WrapAny(&runtimev1alpha1.Module{
				AppName: "rosetta",
			}),
		},
	},
})

func main() {
	var (
		logger            = logger.NewLogger()
		cdc               codec.Codec
		interfaceRegistry codectypes.InterfaceRegistry
	)

	if err := depinject.Inject(Config, &cdc, &interfaceRegistry); err != nil {
		panic(err)
	}

	if err := rosettaCmd.RosettaCommand(interfaceRegistry, cdc).Execute(); err != nil {
		logger.Err(err).Msg("failed to run rosetta")
		os.Exit(1)
	}
}
