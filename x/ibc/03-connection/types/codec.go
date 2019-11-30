package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// SubModuleCdc defines the IBC connection codec.
var SubModuleCdc *codec.Codec

func init() {
	SubModuleCdc = codec.New()
	RegisterCodec(SubModuleCdc)
}

// RegisterCodec registers the IBC connection types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgConnectionOpenInit{}, "ibc/connection/MsgConnectionOpenInit", nil)
	cdc.RegisterConcrete(MsgConnectionOpenTry{}, "ibc/connection/MsgConnectionOpenTry", nil)
	cdc.RegisterConcrete(MsgConnectionOpenAck{}, "ibc/connection/MsgConnectionOpenAck", nil)
	cdc.RegisterConcrete(MsgConnectionOpenConfirm{}, "ibc/connection/MsgConnectionOpenConfirm", nil)
	cdc.RegisterConcrete(ConnectionEnd{}, "ibc/connection/ConnectionEnd", nil)

	SetSubModuleCodec(cdc)
}

func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
