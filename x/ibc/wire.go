package ibc

import (
	wire "github.com/tendermint/go-amino"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(IBCTransferMsg{}, "cosmos-sdk/IBCTransferMsg", nil)
	cdc.RegisterConcrete(IBCReceiveMsg{}, "cosmos-sdk/IBCReceiveMsg", nil)
	cdc.RegisterConcrete(IBCPacket{}, "cosmos-sdk/IBCPacket", nil)
}
