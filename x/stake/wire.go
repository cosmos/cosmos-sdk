package stake

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreateValidator{}, "cosmos-sdk/MsgCreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "cosmos-sdk/MsgEditValidator", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "cosmos-sdk/MsgDelegate", nil)
	cdc.RegisterConcrete(MsgBeginUnbonding{}, "cosmos-sdk/BeginUnbonding", nil)
	cdc.RegisterConcrete(MsgCompleteUnbonding{}, "cosmos-sdk/CompleteUnbonding", nil)
	cdc.RegisterConcrete(MsgBeginRedelegate{}, "cosmos-sdk/BeginRedelegate", nil)
	cdc.RegisterConcrete(MsgCompleteRedelegate{}, "cosmos-sdk/CompleteRedelegate", nil)
}

var msgCdc = wire.NewCodec()

func init() {
	RegisterWire(msgCdc)
	wire.RegisterCrypto(msgCdc)
}
