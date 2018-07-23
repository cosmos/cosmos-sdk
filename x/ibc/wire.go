package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgSend{}, "cosmos-sdk/ibc/Send", nil)
	cdc.RegisterConcrete(MsgReceive{}, "cosmos-sdk/ibc/Receive", nil)
	cdc.RegisterConcrete(MsgCleanup{}, "cosmos-sdk/ibc/Cleanup", nil)
	cdc.RegisterConcrete(MsgOpenConnection{}, "cosmos-sdk/ibc/OpenConnection", nil)
	cdc.RegisterConcrete(MsgUpdateConnection{}, "cosmos-sdk/ibc/UpdateConnection", nil)

	cdc.RegisterConcrete(Datagram{}, "cosmos-sdk/ibc/Datagram", nil)
	cdc.RegisterInterface((*Payload)(nil), nil)
}
