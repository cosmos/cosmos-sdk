package bank

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgSend{}, "cosmos-sdk/Send", nil)
	cdc.RegisterConcrete(MsgIssue{}, "cosmos-sdk/Issue", nil)
	cdc.RegisterConcrete(IBCSendMsg{}, "cosmos-sdk/IBCSend", nil)

	cdc.RegisterConcrete(SendPayload{}, "cosmos-sdk/payload/Send", nil)
	cdc.RegisterConcrete(SendFailReceipt{}, "cosmos-sdk/payload/SendFail", nil)
}
