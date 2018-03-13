package stake

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// XXX complete
func RegisterWire(cdc *wire.Codec) {
	// TODO include option to always include prefix bytes.
	//cdc.RegisterConcrete(SendMsg{}, "cosmos-sdk/SendMsg", nil)
	//cdc.RegisterConcrete(IssueMsg{}, "cosmos-sdk/IssueMsg", nil)
}
