package keys

import (
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

var cdc = wire.NewCodec()

func init() {
	crypto.RegisterWire(cdc)
}
