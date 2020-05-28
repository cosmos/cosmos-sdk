package multisig

import (
	"github.com/tendermint/tendermint/crypto"

	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type GetSignBytesFunc func(mode types.SignMode) ([]byte, error)

type MultisigPubKey interface {
	crypto.PubKey

	VerifyMultisignature(getSignBytes GetSignBytesFunc, sig *types.MultiSignature) bool
	GetPubKeys() []crypto.PubKey
	Threshold() uint32
}
