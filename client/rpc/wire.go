package rpc

import (
	"github.com/cosmos/cosmos-sdk/wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var cdc *wire.Codec

func init() {
	cdc = wire.NewCodec()
	RegisterWire(cdc)
}

func RegisterWire(cdc *wire.Codec) {
	ctypes.RegisterAmino(cdc)
}
