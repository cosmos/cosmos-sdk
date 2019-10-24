package evidence

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/evidence/client"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/cli"
	"github.com/cosmos/cosmos-sdk/x/evidence/client/rest"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var (
	// _ module.AppModule           = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
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
	// TODO: Return proper default genesis state.
	return []byte("[]")
}

// ValidateGenesis performs genesis state validation for the evidence module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	// TODO: Validate genesis state.
	return nil
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
