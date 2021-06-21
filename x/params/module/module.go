package module

import (
	"go.uber.org/dig"

	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/app/compat"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	types2 "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/types"
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

	Handler          app.Handler         `group:"tx"`
	GovRoute         types2.Route        `group:"cosmos.gov.v1.Route"`
	Keeper           paramskeeper.Keeper `security-role:"admin"`
	SubspaceProvider types.SubspaceProvider
}

func (m Module) Provision(key app.ModuleKey) container.Option {
	return container.Provide(func(inputs Inputs) Outputs {
		keeper := paramskeeper.NewKeeper(inputs.Codec, inputs.LegacyAmino, inputs.KeyProvider(key), inputs.TransientKeyProvider(key))
		appMod := params.NewAppModule(keeper)

		return Outputs{
			Handler: compat.AppModuleHandler(key.ID(), appMod),
			Keeper:  keeper,
			SubspaceProvider: func(key app.ModuleKey) types.Subspace {
				return keeper.Subspace(key.ID().Name())
			},
			GovRoute: types2.Route{
				Path:    paramproposal.RouterKey,
				Handler: params.NewParamChangeProposalHandler(keeper),
			},
		}
	})
}
