package rpc

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

var cdc = codec.New()

func init() {
	ctypes.RegisterAmino(cdc.Codec)
}
