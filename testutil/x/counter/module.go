package counter

import (
	"fmt"
	"maps"
	"slices"

	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter/keeper"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
)

var _ appmodule.AppModule = AppModule{}

// AppModule implements an application module
type AppModule struct {
	keeper *keeper.Keeper
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// Name returns the module's name.
//
// Deprecated: kept for legacy reasons.
func (AppModule) Name() string { return types.ModuleName }

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registrar codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registrar)
}

func InvokeSetHooks(keeper *keeper.Keeper, counterHooks map[string]types.CounterHooksWrapper) error {
	if keeper == nil {
		return fmt.Errorf("keeper is nil")
	}
	if counterHooks == nil {
		return fmt.Errorf("counterHooks is nil")
	}

	// Default ordering is lexical by module name.
	// Explicit ordering can be added to the module config if required.
	modNames := slices.Sorted(maps.Keys(counterHooks))
	var multiHooks types.MultiCounterHooks
	for _, modName := range modNames {
		hook, ok := counterHooks[modName]
		if !ok {
			return fmt.Errorf("can't find hooks for module %s", modName)
		}
		multiHooks = append(multiHooks, hook)
	}

	keeper.SetHooks(multiHooks)
	return nil
}
