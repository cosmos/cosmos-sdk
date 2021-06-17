package mock

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

const (
	ModuleName = "mock"
)

var (
	MockAcknowledgement = []byte("mock acknowledgement")
	MockCommitment      = []byte("mock packet commitment")
)

// AppModuleBasic is the mock AppModuleBasic.
type AppModuleBasic struct{}

// Name implements AppModuleBasic interface.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterLegacyAminoCodec implements AppModuleBasic interface.
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces implements AppModuleBasic interface.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// DefaultGenesis implements AppModuleBasic interface.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}

// ValidateGenesis implements the AppModuleBasic interface.
func (AppModuleBasic) ValidateGenesis(codec.JSONMarshaler, client.TxEncodingConfig, json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes implements AppModuleBasic interface.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes implements AppModuleBasic interface.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// GetTxCmd implements AppModuleBasic interface.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd implements AppModuleBasic interface.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule represents the AppModule for the mock module.
type AppModule struct {
	AppModuleBasic
	scopedKeeper capabilitykeeper.ScopedKeeper
}

// NewAppModule returns a mock AppModule instance.
func NewAppModule(sk capabilitykeeper.ScopedKeeper) AppModule {
	return AppModule{
		scopedKeeper: sk,
	}
}

// RegisterInvariants implements the AppModule interface.
func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route implements the AppModule interface.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(ModuleName, nil)
}

// QuerierRoute implements the AppModule interface.
func (AppModule) QuerierRoute() string {
	return ""
}

// LegacyQuerierHandler implements the AppModule interface.
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices implements the AppModule interface.
func (am AppModule) RegisterServices(module.Configurator) {}

// InitGenesis implements the AppModule interface.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis implements the AppModule interface.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	return nil
}

// BeginBlock implements the AppModule interface
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
}

// EndBlock implements the AppModule interface
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ____________________________________________________________________________

// OnChanOpenInit implements the IBCModule interface.
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context, _ channeltypes.Order, _ []string, portID string,
	channelID string, chanCap *capabilitytypes.Capability, _ channeltypes.Counterparty, _ string,
) error {
	// Claim channel capability passed back by IBC module
	if err := am.scopedKeeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}

	return nil
}

// OnChanOpenTry implements the IBCModule interface.
func (am AppModule) OnChanOpenTry(
	ctx sdk.Context, _ channeltypes.Order, _ []string, portID string,
	channelID string, chanCap *capabilitytypes.Capability, _ channeltypes.Counterparty, _, _ string,
) error {
	// Claim channel capability passed back by IBC module
	if err := am.scopedKeeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}

	return nil
}

// OnChanOpenAck implements the IBCModule interface.
func (am AppModule) OnChanOpenAck(sdk.Context, string, string, string) error {
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface.
func (am AppModule) OnChanOpenConfirm(sdk.Context, string, string) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface.
func (am AppModule) OnChanCloseInit(sdk.Context, string, string) error {
	return nil
}

// OnChanCloseConfirm implements the IBCModule interface.
func (am AppModule) OnChanCloseConfirm(sdk.Context, string, string) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface.
func (am AppModule) OnRecvPacket(sdk.Context, channeltypes.Packet) (*sdk.Result, []byte, error) {
	return nil, MockAcknowledgement, nil
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (am AppModule) OnAcknowledgementPacket(sdk.Context, channeltypes.Packet, []byte) (*sdk.Result, error) {
	return nil, nil
}

// OnTimeoutPacket implements the IBCModule interface.
func (am AppModule) OnTimeoutPacket(sdk.Context, channeltypes.Packet) (*sdk.Result, error) {
	return nil, nil
}
