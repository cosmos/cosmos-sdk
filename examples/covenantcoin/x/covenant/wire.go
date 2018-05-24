package covenant

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreateCovenant{}, "covenant/create", nil)
	cdc.RegisterConcrete(MsgSettleCovenant{}, "covenant/settle", nil)
}
