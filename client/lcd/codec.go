package lcd

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

var cdc codec.Codec

func init() {
	aminocdc := codec.New()
	ctypes.RegisterAmino(aminocdc.Codec)
	cdc = aminocdc
}
