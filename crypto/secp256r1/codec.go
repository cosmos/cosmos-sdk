package secp256r1

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
)

const (
	PrivKeyAminoName = "cosmos-sdk/PrivKeySr25519"
	PubKeyAminoName  = "cosmos-sdk/PubKeySr25519"

	// SignatureSize is the size of an Edwards25519 signature. Namely the size of a compressed
	// Sr25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 64
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(PrivKeyNistp256{},
		PubKeyAminoName, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(PubKeyNistp256{},
		PrivKeyAminoName, nil)
}
