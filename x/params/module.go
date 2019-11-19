package params

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the params module.
type AppModuleBasic struct{}

// Name returns the params module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the params module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the params
// module.
func (AppModuleBasic) DefaultGenesis() json.RawMessage { return nil }

// ValidateGenesis performs genesis state validation for the params module.
func (AppModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

// RegisterRESTRoutes registers the REST routes for the params module.
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// GetTxCmd returns no root tx command for the params module.
func (AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the params module.
func (AppModuleBasic) GetQueryCmd(_ *codec.Codec) *cobra.Command { return nil }
