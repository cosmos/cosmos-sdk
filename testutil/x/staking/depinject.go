package staking

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"

	modulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
		appconfig.Invoke(InvokeSetStakingHooks),
	)
}

type ModuleInputs struct {
	depinject.In

	Config                *modulev1.Module
	ValidatorAddressCodec address.ValidatorAddressCodec
	ConsensusAddressCodec address.ConsensusAddressCodec
	AccountKeeper         types.AccountKeeper
	BankKeeper            types.BankKeeper
	Cdc                   codec.Codec
	Environment           appmodule.Environment
	CometInfoService      comet.Service
}

// Dependency Injection Outputs
type ModuleOutputs struct {
	depinject.Out

	StakingKeeper *keeper.Keeper
	Module        appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.AccountKeeper,
		in.BankKeeper,
		in.ValidatorAddressCodec,
		in.ConsensusAddressCodec,
		in.CometInfoService,
	)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper)
	return ModuleOutputs{StakingKeeper: k, Module: m}
}

func InvokeSetStakingHooks(
	config *modulev1.Module,
	keeper *keeper.Keeper,
	stakingHooks map[string]types.StakingHooksWrapper,
) error {
	// all arguments to invokers are optional
	if keeper == nil || config == nil {
		return nil
	}

	modNames := maps.Keys(stakingHooks)
	order := config.HooksOrder
	if len(order) == 0 {
		order = modNames
		sort.Strings(order)
	}

	if len(order) != len(modNames) {
		return fmt.Errorf("len(hooks_order: %v) != len(hooks modules: %v)", order, modNames)
	}

	if len(modNames) == 0 {
		return nil
	}

	var multiHooks types.MultiStakingHooks
	for _, modName := range order {
		hook, ok := stakingHooks[modName]
		if !ok {
			return fmt.Errorf("can't find staking hooks for module %s", modName)
		}

		multiHooks = append(multiHooks, hook)
	}

	keeper.SetHooks(multiHooks)
	return nil
}
