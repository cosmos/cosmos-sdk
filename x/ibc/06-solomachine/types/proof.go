package types

import (
	"github.com/tendermint/tendermint/crypto"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CheckSignature verifies if the the provided public key generated the signature
// over the given data.
func CheckSignature(pubKey crypto.PubKey, data, signature []byte) error {
	if !pubKey.VerifyBytes(data, signature) {
		return sdkerrors.Wrap(ErrSignatureVerificationFailed, "signature verification failed")
	}

	return nil
}
