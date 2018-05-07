package stake

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgDeclareCandidacy{}, "cosmos-sdk/MsgDeclareCandidacy", nil)
	cdc.RegisterConcrete(MsgEditCandidacy{}, "cosmos-sdk/MsgEditCandidacy", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "cosmos-sdk/MsgDelegate", nil)
	cdc.RegisterConcrete(MsgUnbond{}, "cosmos-sdk/MsgUnbond", nil)
}
