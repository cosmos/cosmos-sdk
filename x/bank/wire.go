package bank

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	// TODO: bring this back ...
	/*
		// TODO include option to always include prefix bytes.
		cdc.RegisterConcrete(SendMsg{}, "cosmos-sdk/SendMsg", nil)
		cdc.RegisterConcrete(IssueMsg{}, "cosmos-sdk/IssueMsg", nil)
	*/
}
