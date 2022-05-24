package signing

import (
	cryptotypes "github.com/Stride-Labs/cosmos-sdk/crypto/types"
	"github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/Stride-Labs/cosmos-sdk/types/tx/signing"
)

// SigVerifiableTx defines a transaction interface for all signature verification
// handlers.
type SigVerifiableTx interface {
	types.Tx
	GetSigners() []types.AccAddress
	GetPubKeys() ([]cryptotypes.PubKey, error) // If signer already has pubkey in context, this list will have nil in its place
	GetSignaturesV2() ([]signing.SignatureV2, error)
}

// Tx defines a transaction interface that supports all standard message, signature
// fee, memo, and auxiliary interfaces.
type Tx interface {
	SigVerifiableTx

	types.TxWithMemo
	types.FeeTx
	types.TxWithTimeoutHeight
}
