package evidence

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/evidence/client"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/cli"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/rest"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}

	// TODO: Enable simulation once concrete types are defined.
	// _ module.AppModuleSimulation = AppModuleSimulation{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the evidence module.
type AppModuleBasic struct {
	evidenceHandlers []client.EvidenceHandler // client evidence submission handlers
}

func NewAppModuleBasic(evidenceHandlers ...client.EvidenceHandler) AppModuleBasic {
	return AppModuleBasic{
		evidenceHandlers: evidenceHandlers,
	}
}

// Name returns the evidence module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the evidence module's types to the provided codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// DefaultGenesis returns the evidence module's default genesis state.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the evidence module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var gs GenesisState
	if err := ModuleCdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
	}

	return gs.Validate()
}

// RegisterRESTRoutes registers the evidence module's REST service handlers.
func (a AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	evidenceRESTHandlers := make([]rest.EvidenceRESTHandler, len(a.evidenceHandlers))

	for i, evidenceHandler := range a.evidenceHandlers {
		evidenceRESTHandlers[i] = evidenceHandler.RESTHandler(ctx)
	}

	rest.RegisterRoutes(ctx, rtr, evidenceRESTHandlers)
}

// GetTxCmd returns the evidence module's root tx command.
func (a AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	evidenceCLIHandlers := make([]*cobra.Command, len(a.evidenceHandlers))

	for i, evidenceHandler := range a.evidenceHandlers {
		evidenceCLIHandlers[i] = evidenceHandler.CLIHandler(cdc)
	}

	return cli.GetTxCmd(StoreKey, cdc, evidenceCLIHandlers)
}

// GetTxCmd returns the evidence module's root query command.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(StoreKey, cdc)
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the evidence module.
type AppModule struct {
	AppModuleBasic

	keeper Keeper
}

func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
	}
}

// Name returns the evidence module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// Route returns the evidence module's message routing key.
func (AppModule) Route() string {
	return RouterKey
}

// QuerierRoute returns the evidence module's query routing key.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewHandler returns the evidence module's message Handler.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// NewQuerierHandler returns the evidence module's Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// RegisterInvariants registers the evidence module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// InitGenesis performs the evidence module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, bz json.RawMessage) []abci.ValidatorUpdate {
	var gs GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %s", ModuleName, err))
	}

	InitGenesis(ctx, am.keeper, gs)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the evidence module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return ModuleCdc.MustMarshalJSON(ExportGenesis(ctx, am.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic respective to the evidence module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// EndBlock executes all ABCI EndBlock logic respective to the evidence module. It
// returns no validator updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
