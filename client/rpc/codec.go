package rpc

import (
	"github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}
