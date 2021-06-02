package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

var (
	_ app.Provisioner = &Module{}
)

type Inputs struct {
	container.StructArgs

	Codec        codec.BinaryCodec
	LegacyAmino  *codec.LegacyAmino
	Key          *sdk.KVStoreKey
	TransientKey *sdk.TransientStoreKey
	GovRouter    govtypes.Router
}

type Outputs struct {
	container.StructArgs

	Keeper paramskeeper.Keeper `security-role:"admin"`
}

func (m Module) NewAppHandler(inputs Inputs) (module.AppModule, Outputs, error) {
	keeper := paramskeeper.NewKeeper(inputs.Codec, inputs.LegacyAmino, inputs.Key, inputs.TransientKey)
	appMod := params.NewAppModule(keeper)

	inputs.GovRouter.AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(keeper))

	return appMod, Outputs{Keeper: keeper}, nil
}

func (m Module) Provision(registrar container.Registrar) error {
	return registrar.Provide(func(scope container.Scope, keeper paramskeeper.Keeper) types.Subspace {
		return keeper.Subspace(string(scope))
	})
}
