package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	_ app.Provisioner = &Module{}
)

func (m *Module) Provision(registrar container.Registrar) error {
	return registrar.Provide(func() govtypes.Router {
		return govtypes.NewRouter()
	})
}

type Inputs struct {
	container.StructArgs

	Router     govtypes.Router
	Codec      codec.Codec
	Key        *sdk.KVStoreKey
	ParamStore paramtypes.Subspace

	// TODO: use keepers defined in their respective modules
	AuthKeeper    govtypes.AccountKeeper
	BankKeeper    govtypes.BankKeeper
	StakingKeeper govtypes.StakingKeeper
}

type Outputs struct {
	container.StructArgs
}

func (m *Module) NewAppHandler(inputs Inputs) (module.AppModule, Outputs) {
	k := govkeeper.NewKeeper(
		inputs.Codec,
		inputs.Key,
		inputs.ParamStore,
		inputs.AuthKeeper,
		inputs.BankKeeper,
		inputs.StakingKeeper,
		inputs.Router,
	)

	am := gov.NewAppModule(inputs.Codec, k, inputs.AuthKeeper, inputs.BankKeeper)

	return am, Outputs{}
}
