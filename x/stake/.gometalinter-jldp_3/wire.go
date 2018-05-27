package stake

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreateValidator{}, "cosmos-sdk/MsgCreateValidator", nil)
	cdc.RegisterConcrete(MsgEditValidator{}, "cosmos-sdk/MsgEditValidator", nil)
	cdc.RegisterConcrete(MsgDelegate{}, "cosmos-sdk/MsgDelegate", nil)
	cdc.RegisterConcrete(MsgUnbond{}, "cosmos-sdk/MsgUnbond", nil)
}

var msgCdc = wire.NewCodec()

func init() {
	RegisterWire(msgCdc)
	wire.RegisterCrypto(msgCdc)
}
