package types

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

// generic sealed codec to be used throughout sdk
var MsgCdc *wire.Codec

func init() {
	cdc := wire.NewCodec()
	RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	MsgCdc = cdc
	//MsgCdc = cdc.Seal() //TODO use when upgraded to go-amino 0.9.10
}
