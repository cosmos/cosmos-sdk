package bank

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	// TODO include option to always include prefix bytes.
	//cdc.RegisterConcrete(SendMsg{}, "github.com/cosmos/cosmos-sdk/bank/SendMsg", nil)
	//cdc.RegisterConcrete(IssueMsg{}, "github.com/cosmos/cosmos-sdk/bank/IssueMsg", nil)
}
