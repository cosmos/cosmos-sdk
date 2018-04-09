package server

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc *wire.Codec

func init() {
	cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc)
}
