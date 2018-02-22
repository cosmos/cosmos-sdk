package keys

import (
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

var cdc *wire.Codec

func init() {
	cdc = wire.NewCodec()
	crypto.RegisterWire(cdc)
}

func MarshalJSON(o interface{}) ([]byte, error) {
	return cdc.MarshalJSON(o)
}
