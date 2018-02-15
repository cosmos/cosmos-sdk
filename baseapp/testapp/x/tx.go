package baseapp

import (
	"github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// testing transaction
type TestTx struct {
	sdk.Msg
}

// nolint
func (tx TestTx) GetMsg() sdk.Msg                   { return tx.Msg }
func (tx TestTx) GetSigners() []crypto.Address      { return nil }
func (tx TestTx) GetFeePayer() crypto.Address       { return nil }
func (tx TestTx) GetSignatures() []sdk.StdSignature { return nil }
func IsTestAppTx(tx sdk.Tx) bool {
	_, ok := tx.(TestTx)
	return ok
}
