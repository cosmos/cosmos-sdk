package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func createCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}
