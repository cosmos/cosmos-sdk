package wire

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-crypto"
)

type Codec = amino.Codec

func NewCodec() *Codec {
	cdc := amino.NewCodec()
	return cdc
}

func RegisterCrypto(cdc *Codec) {
	crypto.RegisterAmino(cdc)
}
