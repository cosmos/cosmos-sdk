package bank

import (
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/spf13/viper"

	modulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/auth/ante"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagMinGasPricesV2 = "server.minimum-gas-prices"
	feegrantModuleName = "feegrant"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetSendRestrictions, InvokeCheckFeeGrantPresent),
	)
}

type ModuleInputs struct {
	depinject.In

	Config      *modulev1.Module
	Cdc         codec.Codec
	Environment appmodule.Environment

	AccountKeeper types.AccountKeeper
	Viper         *viper.Viper `optional:"true"` // server v2
}

type ModuleOutputs struct {
	depinject.Out

	BankKeeper     keeper.BaseKeeper
	Module         appmodule.AppModule
	FeeTxValidator ante.FeeTxValidator // pass deduct fee decorator to feegrant TxValidator
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// Configure blocked module accounts.
	//
	// Default behavior for blockedAddresses is to regard any module mentioned in
	// AccountKeeper's module account permissions as blocked.
	blockedAddresses := make(map[string]bool)
	if len(in.Config.BlockedModuleAccountsOverride) > 0 {
		for _, moduleName := range in.Config.BlockedModuleAccountsOverride {
			addrStr, err := in.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(moduleName))
			if err != nil {
				panic(err)
			}
			blockedAddresses[addrStr] = true
		}
	} else {
		for _, permission := range in.AccountKeeper.GetModulePermissions() {
			addrStr, err := in.AccountKeeper.AddressCodec().BytesToString(permission.GetAddress())
			if err != nil {
				panic(err)
			}
			blockedAddresses[addrStr] = true
		}
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	authStr, err := in.AccountKeeper.AddressCodec().BytesToString(authority)
	if err != nil {
		panic(err)
	}

	bankKeeper := keeper.NewBaseKeeper(
		in.Environment,
		in.Cdc,
		in.AccountKeeper,
		blockedAddresses,
		authStr,
	)
	m := NewAppModule(in.Cdc, bankKeeper, in.AccountKeeper)

	var minGasPrices sdk.DecCoins
	if in.Viper != nil {
		minGasPricesStr := in.Viper.GetString(FlagMinGasPricesV2)
		minGasPrices, err = sdk.ParseDecCoins(minGasPricesStr)
		if err != nil {
			panic(fmt.Sprintf("invalid minimum gas prices: %v", err))
		}
	}

	var feeTxValidator ante.FeeTxValidator
	if in.AccountKeeper != nil {
		feeTxValidator = ante.NewDeductFeeDecorator(in.AccountKeeper, bankKeeper, nil, nil)
		// set min gas price in deduct fee decorator
		feeTxValidator.SetMinGasPrices(minGasPrices)
		// pass deduct fee decorator to app module
		m.SetFeeTxValidator(feeTxValidator)
	}

	return ModuleOutputs{BankKeeper: bankKeeper, Module: m, FeeTxValidator: feeTxValidator}
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

// TODO: Remove below check and move deduct-fee-decorator check completely to x/bank txValidator once sims v2 is done.
// Sims v2 PR will remove dependency between x/bank and x/feegrant.https://github.com/cosmos/cosmos-sdk/pull/20940
func InvokeCheckFeeGrantPresent(modules map[string]appmodule.AppModule) error {
	_, ok := modules[feegrantModuleName]
	if ok {
		// get bank module
		bankMod, ok := modules[types.ModuleName]
		if !ok {
			return nil
		}

		// set isFeegrant
		m := bankMod.(AppModule)
		m.SetFeeTxValidator(nil)
	}
	return nil
}
