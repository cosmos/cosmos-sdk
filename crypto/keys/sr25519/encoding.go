package sr25519

import (
	"github.com/tendermint/tendermint/crypto"
)

var _ crypto.PrivKey = PrivKey{}

const (
	PrivKeyName = "cosmos-sdk/PrivKeySr25519"
	PubKeyName  = "cosmos-sdk/PubKeySr25519"

	// SignatureSize is the size of an Edwards25519 signature. Namely the size of a compressed
	// Sr25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 64
)
