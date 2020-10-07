package changepubkey

import (
	"encoding/json"

	"github.com/gogo/protobuf/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the sub-pubkey
// module. The module itself contain no special logic or state other than message
// handling.
type AppModuleBasic struct{}

// Name returns the module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types with the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interfaces and implementations with
// the given interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns the module's default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return []byte("{}")
}

// ValidateGenesis performs genesis state validation. Currently, this is a no-op.
func (AppModuleBasic) ValidateGenesis(_ codec.JSONMarshaler, _ client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes registers module's REST handlers. Currently, this is a no-op.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCRoutes registers the module's gRPC Gateway routes. Currently, this
// is a no-op.
func (a AppModuleBasic) RegisterGRPCRoutes(_ client.Context, _ *runtime.ServeMux) {}

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

	accountKeeper keeper.AccountKeeper
}

// NewAppModule returns an instance of AppModule
func NewAppModule(ak keeper.AccountKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  ak,
	}
}

// RegisterInvariants performs a no-op; there are no invariants to enforce.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the module's message router and handler.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.accountKeeper))
}

// QuerierRoute returns querier route for changepubkey
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterQueryService performs a no-op.
func (am AppModule) RegisterQueryService(server grpc.Server) {
}

// LegacyQuerierHandler performs a no-op.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs a no-op.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// BeginBlock performs a no-op.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs a no-op.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis is always empty, as InitGenesis does nothing either.
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return []byte("{}")
}
