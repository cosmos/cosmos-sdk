package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var SubModuleCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgConnectionOpenInit{}, "ibc/connection/MsgConnectionOpenInit", nil)
	cdc.RegisterConcrete(MsgConnectionOpenTry{}, "ibc/connection/MsgConnectionOpenTry", nil)
	cdc.RegisterConcrete(MsgConnectionOpenAck{}, "ibc/connection/MsgConnectionOpenAck", nil)
	cdc.RegisterConcrete(MsgConnectionOpenConfirm{}, "ibc/connection/MsgConnectionOpenConfirm", nil)
}

func SetMsgConnectionCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
