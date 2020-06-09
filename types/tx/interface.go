package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtoTx interface {
	sdk.Tx
	FeeTx
	TxWithMemo
	SigTx

	GetBody() *TxBody
	GetAuthInfo() *AuthInfo
	GetSignatures() [][]byte

	GetBodyBytes() []byte
	GetAuthInfoBytes() []byte
}

// FeeTx defines the interface to be implemented by Tx to use the FeeDecorators
type FeeTx interface {
	sdk.Tx
	GetGas() uint64
	GetFee() sdk.Coins
	FeePayer() sdk.AccAddress
}

// Tx must have GetMemo() method to use ValidateMemoDecorator
type TxWithMemo interface {
	sdk.Tx
	GetMemo() string
}

type SigTx interface {
	sdk.Tx
	GetSignaturesV2() ([]SignatureV2, error)
}
