package types

import (
	"github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SigVerifiableTx defines a Tx interface for all signature verification decorators
type SigVerifiableTx interface {
	types.Tx
	txtypes.SigTx
	GetSignBytes(ctx types.Context, acc types2.AccountI) []byte
}

