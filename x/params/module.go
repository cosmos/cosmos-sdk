package params

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/client/rest"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	_ sdk.AppModuleBasic = AppModuleBasic{}
)

const moduleName = "params"

// app module basics object
type AppModuleBasic struct{}

// module name
func (AppModuleBasic) Name() string {
	return moduleName
}

// register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage { return nil }

// module validate genesis
func (AppModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router, cdc *codec.Codec) {
	rest.RegisterRoutes(ctx, rtr, cdc, StoreKey)
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }
