package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/compat"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	parammodule "github.com/cosmos/cosmos-sdk/x/params/module"
	"go.uber.org/dig"
)

var (
	_ app.TypeProvider = &Module{}
	_ app.Provisioner  = &Module{}
)

func (m *Module) RegisterTypes(registry types.InterfaceRegistry) {
	govtypes.RegisterInterfaces(registry)
}

type Route struct {
	Path    string
	Handler govtypes.Handler
}

type Inputs struct {
	dig.In

	Codec            codec.Codec
	KeyProvider      app.KVStoreKeyProvider
	SubspaceProvider parammodule.SubspaceProvider
	Routes           []Route `group:"cosmos.gov.v1.Route"`

	// TODO: use keepers defined in their respective modules
	AuthKeeper    govtypes.AccountKeeper
	BankKeeper    govtypes.BankKeeper
	StakingKeeper govtypes.StakingKeeper
}

type Outputs struct {
	dig.Out

	Handler app.Handler `group:"app.handler"`
}

func (m *Module) Provision(key app.ModuleKey, registrar container.Registrar) error {
	err := registrar.Provide(func(inputs Inputs) Outputs {
		router := govtypes.NewRouter()
		for _, route := range inputs.Routes {
			router.AddRoute(route.Path, route.Handler)
		}
		k := govkeeper.NewKeeper(
			inputs.Codec,
			inputs.KeyProvider(key),
			inputs.SubspaceProvider(key),
			inputs.AuthKeeper,
			inputs.BankKeeper,
			inputs.StakingKeeper,
			router,
		)

		am := gov.NewAppModule(inputs.Codec, k, inputs.AuthKeeper, inputs.BankKeeper)

		return Outputs{
			Handler: compat.AppModuleHandler(am),
		}
	})

	if err != nil {
		return err
	}

	return registrar.Provide(func() govtypes.Router {
		return govtypes.NewRouter()
	})
}
