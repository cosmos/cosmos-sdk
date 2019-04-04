package types

import (
	"github.com/spf13/cobra"
)

// ModuleClients helps modules provide a standard interface for exporting client functionality
type ModuleClients interface {
	GetQueryCmd() *cobra.Command
	GetTxCmd() *cobra.Command
}

//// AppModule is the standard form for an application module
//type AppModule interface {

//// registers
//RegisterCodec(*codec.Codec)
//RegisterInvariants(CrisisKeeper)

//// routes
//MessageRoute()
//NewMessageHandler()
//QuerierRoute()
//NewQuerierHandler()

//// genesis
//DefaultGenesisState() json.RawMessage
//ValidateGenesis(json.RawMessage) error
//ProcessGenesis(sdk.Context, json.RawMessage) ([]abci.ValidatorUpdate, error)
//ExportGenesis(sdk.Context) json.RawMessage
//}
