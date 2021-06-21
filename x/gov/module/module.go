package module

import (
	"github.com/spf13/cobra"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/app/compat"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	types2 "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	_ app.TypeProvider = &Module{}
	_ app.Provisioner  = &Module{}
)

func (m *Module) RegisterTypes(registry types.InterfaceRegistry) {
	govtypes.RegisterInterfaces(registry)
}

type inputs struct {
	container.In

	Codec            codec.Codec
	KeyProvider      app.KVStoreKeyProvider
	SubspaceProvider types2.SubspaceProvider
	Routes           []govtypes.Route `group:"cosmos.gov.v1.Route"`

	// TODO: use keepers defined in their respective modules
	AuthKeeper    govtypes.AccountKeeper
	BankKeeper    govtypes.BankKeeper
	StakingKeeper govtypes.StakingKeeper
}

type outputs struct {
	container.Out

	Handler app.Handler `group:"tx"`
}

type cliCommands struct {
	dig.Out

	TxCmd    *cobra.Command `group:"tx"`
	QueryCmd *cobra.Command `group:"query"`
}

func (m *Module) Provision(key app.ModuleKey) container.Option {
	return container.Provide(
		func(inputs inputs) outputs {
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

			return outputs{
				Handler: compat.AppModuleHandler(key.ID(), am),
			}
		},
		func() cliCommands {
			amb := gov.AppModuleBasic{}
			return cliCommands{
				TxCmd:    amb.GetTxCmd(),
				QueryCmd: amb.GetQueryCmd(),
			}
		},
	)
}
