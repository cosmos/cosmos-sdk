package types

import (
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

// ModuleClients helps modules provide a standard interface for exporting client functionality
type ModuleClients interface {
	GetQueryCmd(storeKey string, cdc *amino.Codec) *cobra.Command
	GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command
}
