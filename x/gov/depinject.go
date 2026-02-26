package gov

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	modulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule, ProvideKeyTable),
		appmodule.Invoke(InvokeAddRoutes, InvokeSetHooks))
}

type ModuleInputs struct {
	depinject.In

	Config           *modulev1.Module
	Cdc              codec.Codec
	StoreService     store.KVStoreService
	ModuleKey        depinject.OwnModuleKey
	MsgServiceRouter baseapp.MessageRouter

	AccountKeeper      govtypes.AccountKeeper
	BankKeeper         govtypes.BankKeeper
	StakingKeeper      govtypes.StakingKeeper
	DistributionKeeper govtypes.DistributionKeeper

	// CustomCalculateVoteResultsAndVotingPowerFn is an optional input to set a custom CalculateVoteResultsAndVotingPowerFn.
	// If this function is not provided, the default function is used.
	CustomCalculateVoteResultsAndVotingPowerFn keeper.CalculateVoteResultsAndVotingPowerFn `optional:"true"`

	// LegacySubspace is used solely for migration of x/params managed parameters
	LegacySubspace govtypes.ParamSubspace `optional:"true"`
}

type ModuleOutputs struct {
	depinject.Out

	Module       appmodule.AppModule
	Keeper       *keeper.Keeper
	HandlerRoute v1beta1.HandlerRoute
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	defaultConfig := govtypes.DefaultConfig()
	if in.Config.MaxMetadataLen != 0 {
		defaultConfig.MaxMetadataLen = in.Config.MaxMetadataLen
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	var opts []keeper.InitOption
	if in.CustomCalculateVoteResultsAndVotingPowerFn != nil {
		opts = append(opts, keeper.WithCustomCalculateVoteResultsAndVotingPowerFn(in.CustomCalculateVoteResultsAndVotingPowerFn))
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.AccountKeeper,
		in.BankKeeper,
		in.StakingKeeper,
		in.DistributionKeeper,
		in.MsgServiceRouter,
		defaultConfig,
		authority.String(),
		opts...,
	)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.LegacySubspace)
	hr := v1beta1.HandlerRoute{Handler: v1beta1.ProposalHandler, RouteKey: govtypes.RouterKey}

	return ModuleOutputs{Module: m, Keeper: k, HandlerRoute: hr}
}

func ProvideKeyTable() paramtypes.KeyTable {
	return v1.ParamKeyTable() //nolint:staticcheck // we still need this for upgrades
}

func InvokeAddRoutes(keeper *keeper.Keeper, routes []v1beta1.HandlerRoute) {
	if keeper == nil || routes == nil {
		return
	}

	// Default route order is a lexical sort by RouteKey.
	// Explicit ordering can be added to the module config if required.
	slices.SortFunc(routes, func(x, y v1beta1.HandlerRoute) int {
		return strings.Compare(x.RouteKey, y.RouteKey)
	})

	router := v1beta1.NewRouter()
	for _, r := range routes {
		router.AddRoute(r.RouteKey, r.Handler)
	}
	keeper.SetLegacyRouter(router)
}

func InvokeSetHooks(keeper *keeper.Keeper, govHooks map[string]govtypes.GovHooksWrapper) error {
	if keeper == nil || govHooks == nil {
		return nil
	}

	// Default ordering is lexical by module name.
	// Explicit ordering can be added to the module config if required.
	modNames := slices.Sorted(maps.Keys(govHooks))
	order := modNames
	sort.Strings(order)

	var multiHooks govtypes.MultiGovHooks
	for _, modName := range order {
		hook, ok := govHooks[modName]
		if !ok {
			return fmt.Errorf("can't find staking hooks for module %s", modName)
		}
		multiHooks = append(multiHooks, hook)
	}

	keeper.SetHooks(multiHooks)
	return nil
}
