package counter

import (
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/counter/keeper"
	"github.com/cosmos/cosmos-sdk/x/counter/types"
)

var (
	_ module.HasName = AppModule{}

	_ appmodule.AppModule             = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
)

// AppModule implements an application module
type AppModule struct {
	keeper keeper.Keeper
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// Name returns the consensus module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	types.RegisterInterfaces(registrar)
}
