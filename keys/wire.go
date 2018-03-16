package keys

import (
	amino "github.com/tendermint/go-amino"
	crypto "github.com/tendermint/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
}
