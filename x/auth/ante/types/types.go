package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FeeTx defines the interface to be implemented by Tx to use the FeeDecorators
type FeeTx interface {
	types.Tx
	GetGas() uint64
	GetFee() types.Coins
	FeePayer() types.AccAddress
}

// Tx must have GetMemo() method to use ValidateMemoDecorator
type TxWithMemo interface {
	types.Tx
	GetMemo() string
}

// SigVerifiableTx defines a Tx interface for all signature verification decorators
type SigVerifiableTx interface {
	types.Tx
	HasPubKeysTx
	GetSignatures() [][]byte
	GetSignBytes(ctx types.Context, acc types2.AccountI) []byte
}

type HasPubKeysTx interface {
	types.Tx
	GetSigners() []types.AccAddress
	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place
}
