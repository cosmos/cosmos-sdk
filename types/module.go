package types

import (
	"encoding/json"

	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ModuleClients helps modules provide a standard interface for exporting client functionality
type ModuleClients interface {
	GetQueryCmd() *cobra.Command
	GetTxCmd() *cobra.Command
}

// ModuleGenesis is the standard form for module genesis functions
type ModuleGenesis interface {
	Preprocess(sdk.Context) ([]abci.ValidatorUpdate, error)
	ValidateInput() error
	ExportGenesis(sdk.Context) json.RawMessage
}
