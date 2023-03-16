package rekeying

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"cosmossdk.io/core/appmodule"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/rekeying/client/cli"
	rekeyingkeeper "github.com/cosmos/cosmos-sdk/x/auth/rekeying/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/rekeying/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the sub-vesting
// module. The module itself contain no special logic or state other than message
// handling.
type AppModuleBasic struct{}

// Name returns the module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers the module's types with the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interfaces and implementations with
// the given interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns the module's default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	return []byte("{}")
}

// ValidateGenesis performs genesis state validation. Currently, this is a no-op.
func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterGRPCGatewayRoutes registers the module's gRPC Gateway routes. Currently, this
// is a no-op.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *gwruntime.ServeMux) {}

// GetTxCmd returns the root tx command for the auth module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the module's root query command. Currently, this is a no-op.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule extends the AppModuleBasic implementation by implementing the
// AppModule interface.
type AppModule struct {
	AppModuleBasic

	accountKeeper  authkeeper.AccountKeeper
	rekeyingKeeper rekeyingkeeper.Keeper
}

func NewAppModule(ak authkeeper.AccountKeeper, rk rekeyingkeeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  ak,
		rekeyingKeeper: rk,
	}
}

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterInvariants performs a no-op; there are no invariants to enforce.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), rekeyingkeeper.NewMsgServerImpl(am.accountKeeper, am.rekeyingKeeper))
}

// InitGenesis performs a no-op.
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis is always empty, as InitGenesis does nothing either.
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
