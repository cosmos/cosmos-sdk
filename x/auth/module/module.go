package module

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type Inputs struct {
	container.StructArgs

	Codec      codec.Codec
	Key        *sdk.KVStoreKey
	ParamStore paramtypes.Subspace
}

type Outputs struct {
	container.StructArgs

	ViewKeeper types.ViewKeeper
	Keeper     types.Keeper `security-role:"admin"`
}

func (m Module) NewAppModule(inputs Inputs) (module.AppModule, Outputs, error) {
	var accCtr types.AccountConstructor
	if m.AccountConstructor != nil {
		err := inputs.Codec.UnpackAny(m.AccountConstructor, &accCtr)
		if err != nil {
			return nil, Outputs{}, err
		}
	} else {
		accCtr = DefaultAccountConstructor{}
	}

	perms := map[string][]string{}
	for _, perm := range m.Permissions {
		perms[perm.Address] = perm.Permissions
	}

	var randomGenesisAccountsProvider types.RandomGenesisAccountsProvider
	if m.RandomGenesisAccountsProvider != nil {
		err := inputs.Codec.UnpackAny(m.RandomGenesisAccountsProvider, &randomGenesisAccountsProvider)
		if err != nil {
			return nil, Outputs{}, err
		}
	} else {
		randomGenesisAccountsProvider = DefaultRandomGenesisAccountsProvider{}
	}

	keeper := authkeeper.NewAccountKeeper(inputs.Codec, inputs.Key, inputs.ParamStore, func() types.AccountI {
		return accCtr.NewAccount()
	}, perms)
	appMod := auth.NewAppModule(inputs.Codec, keeper, func(simState *module.SimulationState) types.GenesisAccounts {
		return randomGenesisAccountsProvider.RandomGenesisAccounts(simState)
	})

	return appMod, Outputs{
		ViewKeeper: viewOnlyKeeper{keeper},
		Keeper:     keeper,
	}, nil
}

// viewOnlyKeeper wraps the full keeper in a view-only interface which can't be easily type cast to the full keeper interface
type viewOnlyKeeper struct {
	k authkeeper.AccountKeeper
}

func (v viewOnlyKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI {
	return v.k.GetAccount(ctx, addr)
}

func (v viewOnlyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return v.k.GetModuleAddress(moduleName)
}

func (v viewOnlyKeeper) ValidatePermissions(macc types.ModuleAccountI) error {
	return v.k.ValidatePermissions(macc)
}

func (v viewOnlyKeeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string) {
	return v.GetModuleAddressAndPermissions(moduleName)
}

func (v viewOnlyKeeper) GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (types.ModuleAccountI, []string) {
	return v.GetModuleAccountAndPermissions(ctx, moduleName)
}

func (v viewOnlyKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI {
	return v.k.GetModuleAccount(ctx, moduleName)
}

var _ types.ViewKeeper = viewOnlyKeeper{}

func (m DefaultAccountConstructor) NewAccount() types.AccountI {
	return &types.BaseAccount{}
}

func (m DefaultRandomGenesisAccountsProvider) RandomGenesisAccounts(simState *module.SimulationState) types.GenesisAccounts {
	return simulation.RandomGenesisAccounts(simState)
}
