package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	//cdc.RegisterConcrete(IBCTransferMsg{}, "github.com/cosmos/cosmos-sdk/x/ibc/IBCTransferMsg", nil)
	//cdc.RegisterConcrete(IBCReceiveMsg{}, "github.com/cosmos/cosmos-sdk/x/ibc/IBCReceiveMsg", nil)
}
