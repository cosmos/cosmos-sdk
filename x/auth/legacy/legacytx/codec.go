package legacytx

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func init() {
	var amino = codec.NewLegacyAmino()
	amino.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx", nil)
}
