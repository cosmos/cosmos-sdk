package signing

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// SigVerifiableTx defines a Tx interface for all signature verification decorators
type SigVerifiableTx interface {
	types.Tx
	GetSigners() []types.AccAddress
	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
	GetSignatures() [][]byte
	GetSignaturesV2() ([]signing.SignatureV2, error)
}

// SigFeeMemoTx defines an interface for transactions that support all standard message, signature,
// fee and memo interfaces.
type SigFeeMemoTx interface {
	SigVerifiableTx
	types.TxWithMemo
	types.FeeTx
}
