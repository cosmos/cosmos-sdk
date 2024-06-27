package types

import (
	cosmoscrypto "github.com/cosmos/crypto/types"
	"github.com/cosmos/gogoproto/proto"
)

type (
	// SdkPrivKey is a private key type parameterized by PubKey, which includes proto functions.
	SdkPrivKey = cosmoscrypto.PrivKey[PubKey]
	// Address is an alias of cosmoscrypto.Address
	Address = cosmoscrypto.Address
)

// PubKey defines a public key and extends proto.Message.
type PubKey interface {
	proto.Message
	cosmoscrypto.PubKey
}

// PrivKey defines a private key and extends proto.Message
type PrivKey interface {
	proto.Message
	SdkPrivKey
}

// LedgerPrivKeyAminoJSON is a PrivKey type that supports signing with
// SIGN_MODE_LEGACY_AMINO_JSON. It is added as a non-breaking change, instead of directly
// on the PrivKey interface (whose Sign method will sign with TEXTUAL),
// and will be deprecated/removed once LEGACY_AMINO_JSON is removed.
type LedgerPrivKeyAminoJSON interface {
	SdkPrivKey
	// SignLedgerAminoJSON signs a messages on the Ledger device using
	// SIGN_MODE_LEGACY_AMINO_JSON.
	SignLedgerAminoJSON(msg []byte) ([]byte, error)
}
