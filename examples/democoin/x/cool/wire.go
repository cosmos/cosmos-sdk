package cool

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgQuiz{}, "cool/Quiz", nil)
	cdc.RegisterConcrete(MsgSetTrend{}, "cool/SetTrend", nil)
}
