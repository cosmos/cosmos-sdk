package bank

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(SendMsg{}, "cosmos-sdk/Send", nil)
	cdc.RegisterConcrete(IssueMsg{}, "cosmos-sdk/Issue", nil)
}
