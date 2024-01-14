package tx

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type SigVerifiableTxV2 interface {
	types.TxV2
	GetSigners() ([][]byte, error)
	GetPubKeys() ([]cryptotypes.PubKey, error) // If signer already has pubkey in context, this list will have nil in its place
	GetSignaturesV2() ([]signing.SignatureV2, error)
}

type TxV2 interface {
	SigVerifiableTxV2

	types.TxV2WithMemo
	types.TxV2WithFee
	types.TxV2WithTimeoutHeight
	types.HasValidateBasic
}
