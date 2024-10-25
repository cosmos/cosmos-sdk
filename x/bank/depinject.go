package bank

import (
	"fmt"
	"maps"
	"slices"
	"sort"

	modulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/moduleaccounts"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetSendRestrictions),
		appconfig.Invoke(InvokeSetDefaultBlockedAddresses),
	)
}

type ModuleInputs struct {
	depinject.In

	Config                *modulev1.Module
	Cdc                   codec.Codec
	Environment           appmodule.Environment
	AddressCodec          address.Codec
	ModuleAccountsService moduleaccounts.ServiceWithPerms

	AccountKeeper types.AccountKeeper
}

type ModuleOutputs struct {
	depinject.Out

	BankKeeper keeper.BaseKeeper
	Module     appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// Configure blocked module accounts.
	//
	// Default behavior for blockedAddresses is to regard any module mentioned in
	// AccountKeeper's module account permissions as blocked.
	blockedAddresses := make(map[string]bool)
	if len(in.Config.BlockedModuleAccountsOverride) > 0 {
		for _, moduleName := range in.Config.BlockedModuleAccountsOverride {
			addrStr, err := in.AddressCodec.BytesToString(authtypes.NewModuleAddress(moduleName))
			if err != nil {
				panic(err)
			}
			blockedAddresses[addrStr] = true
		}
	}

	// default to governance authority if not provided
	// TODO: @facu use module accounts service to get the authority
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	authStr, err := in.AddressCodec.BytesToString(authority)
	if err != nil {
		panic(err)
	}

	bankKeeper := keeper.NewBaseKeeper(
		in.Environment,
		in.Cdc,
		in.AccountKeeper,
		blockedAddresses,
		authStr,
		in.ModuleAccountsService,
	)
	m := NewAppModule(in.Cdc, bankKeeper, in.AccountKeeper)

	return ModuleOutputs{BankKeeper: bankKeeper, Module: m}
}

// InvokeSetDefaultBlockedAddresses only sets the blocked addresses if they are not already set.
func InvokeSetDefaultBlockedAddresses(
	config *modulev1.Module,
	keeper keeper.BaseKeeper,
	moduleAccountsService moduleaccounts.ServiceWithPerms,
	addrCdc address.Codec,
) error {
	if len(keeper.GetBlockedAddresses()) > 0 {
		return nil
	}

	// @facu fix this
	blockedAddresses := make(map[string]bool)
	for _, addr := range moduleAccountsService.AllAccounts() {
		addrStr, err := addrCdc.BytesToString(addr)
		if err != nil {
			return err
		}
		blockedAddresses[addrStr] = true
	}

	keeper.SetBlockedAddresses(blockedAddresses)

	return nil
}

func InvokeSetSendRestrictions(
	config *modulev1.Module,
	keeper keeper.BaseKeeper,
	restrictions map[string]types.SendRestrictionFn,
) error {
	if config == nil {
		return nil
	}

	modules := slices.Collect(maps.Keys(restrictions))
	order := config.RestrictionsOrder
	if len(order) == 0 {
		order = modules
		sort.Strings(order)
	}

	if len(order) != len(modules) {
		return fmt.Errorf("len(restrictions order: %v) != len(restriction modules: %v)", order, modules)
	}

	if len(modules) == 0 {
		return nil
	}

	for _, module := range order {
		restriction, ok := restrictions[module]
		if !ok {
			return fmt.Errorf("can't find send restriction for module %s", module)
		}

		keeper.AppendSendRestriction(restriction)
	}

	return nil
}
