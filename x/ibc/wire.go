package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*Payload)(nil), nil)

	cdc.RegisterConcrete(ReceiveMsg{}, "cosmos-sdk/ReceiveMsg", nil)
	cdc.RegisterConcrete(ReceiptMsg{}, "cosmos-sdk/ReceiptMsg", nil)
	cdc.RegisterConcrete(ReceiveCleanupMsg{}, "cosmos-sdk/ReceiveCleanupMsg", nil)
	cdc.RegisterConcrete(ReceiptCleanupMsg{}, "cosmos-sdk/ReceiptCleanupMsg", nil)
	cdc.RegisterConcrete(OpenChannelMsg{}, "cosmos-sdk/OpenChannelMsg", nil)
	cdc.RegisterConcrete(UpdateChannelMsg{}, "cosmos-sdk/UpdateChannelMsg", nil)

}
