package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
)

// SubModuleCdc defines the IBC connection codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the IBC connection types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.ConnectionI)(nil), nil)
	cdc.RegisterInterface((*exported.CounterpartyI)(nil), nil)
	cdc.RegisterConcrete(ConnectionEnd{}, "ibc/connection/ConnectionEnd", nil)

	cdc.RegisterConcrete(MsgConnectionOpenInit{}, "ibc/connection/MsgConnectionOpenInit", nil)
	cdc.RegisterConcrete(MsgConnectionOpenTry{}, "ibc/connection/MsgConnectionOpenTry", nil)
	cdc.RegisterConcrete(MsgConnectionOpenAck{}, "ibc/connection/MsgConnectionOpenAck", nil)
	cdc.RegisterConcrete(MsgConnectionOpenConfirm{}, "ibc/connection/MsgConnectionOpenConfirm", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc connection codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
