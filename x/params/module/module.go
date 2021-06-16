package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/compat"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	govmodule "github.com/cosmos/cosmos-sdk/x/gov/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"go.uber.org/dig"
)

var (
	_ app.Provisioner = &Module{}
)

type Inputs struct {
	dig.In

	Codec                codec.BinaryCodec
	LegacyAmino          *codec.LegacyAmino
	KeyProvider          app.KVStoreKeyProvider
	TransientKeyProvider app.TransientStoreKeyProvider
}

type Outputs struct {
	dig.Out

	Handler          app.Handler         `group:"app.handler"`
	GovRoute         govmodule.Route     `group:"cosmos.gov.v1.Route"`
	Keeper           paramskeeper.Keeper `security-role:"admin"`
	SubspaceProvider SubspaceProvider
}

type SubspaceProvider func(app.ModuleKey) types.Subspace

func (m Module) Provision(key app.ModuleKey, registrar container.Registrar) error {
	return registrar.Provide(func(inputs Inputs) Outputs {
		keeper := paramskeeper.NewKeeper(inputs.Codec, inputs.LegacyAmino, inputs.KeyProvider(key), inputs.TransientKeyProvider(key))
		appMod := params.NewAppModule(keeper)

		return Outputs{
			Handler: compat.AppModuleHandler(appMod),
			Keeper:  keeper,
			SubspaceProvider: func(key app.ModuleKey) types.Subspace {
				return keeper.Subspace(key.Name())
			},
			GovRoute: govmodule.Route{
				Path:    paramproposal.RouterKey,
				Handler: params.NewParamChangeProposalHandler(keeper),
			},
		}
	})
}
