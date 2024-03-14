package staking

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"

	modulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/simulation"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
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
}

// Dependency Injection Outputs
type ModuleOutputs struct {
	depinject.Out

	StakingKeeper *keeper.Keeper
	Module        appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	as, err := in.AccountKeeper.AddressCodec().BytesToString(authority)
	if err != nil {
		panic(err)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.AccountKeeper,
		in.BankKeeper,
		as,
		in.ValidatorAddressCodec,
		in.ConsensusAddressCodec,
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

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the staking module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for staking module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the staking module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, simState.TxConfig,
		am.accountKeeper, am.bankKeeper, am.keeper,
	)
}
