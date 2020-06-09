package multisig

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type GetSignBytesFunc func(mode signing.SignMode) ([]byte, error)

type MultisigPubKey interface {
	crypto.PubKey

	VerifyMultisignature(getSignBytes GetSignBytesFunc, sig *signing.MultiSignatureData) bool
	GetPubKeys() []crypto.PubKey
	Threshold() uint32
}
