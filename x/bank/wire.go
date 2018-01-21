package bank

import (
	"github.com/tendermint/go-wire"
)

func RegisterWire(cdc *wire.Codec) {
	// TODO include option to always include prefix bytes.
	cdc.RegisterConcrete(SendMsg{}, "cosmos-sdk/SendMsg", nil)
	cdc.RegisterConcrete(IssueMsg{}, "cosmos-sdk/IssueMsg", nil)
}
