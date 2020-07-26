package multisig

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/KiraCore/cosmos-sdk/types/tx/signing"
)

// PubKey defines a type which supports multi-signature verification via MultiSignatureData
// which supports multiple SignMode's.
type PubKey interface {
	crypto.PubKey

	// VerifyMultisignature verifies the provide multi-signature represented by MultiSignatureData
	// using getSignBytes to retrieve the sign bytes to verify against for the provided mode.
	VerifyMultisignature(getSignBytes GetSignBytesFunc, sig *signing.MultiSignatureData) error

	// GetPubKeys returns the crypto.PubKey's nested within the multi-sig PubKey
	GetPubKeys() []crypto.PubKey
}

// GetSignBytesFunc defines a function type which returns sign bytes for a given SignMode or an error.
// It will generally be implemented as a closure which wraps whatever signable object signatures are
// being verified against.
type GetSignBytesFunc func(mode signing.SignMode) ([]byte, error)
