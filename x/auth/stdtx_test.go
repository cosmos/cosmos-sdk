package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// func newStdFee() StdFee {
// 	return NewStdFee(100,
// 		Coin{"atom", 150},
// 	)
// }

func TestStdTx(t *testing.T) {
	priv := crypto.GenPrivKeyEd25519()
	addr := priv.PubKey().Address()
	msg := sdk.NewTestMsg(addr)
	fee := newStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msg, fee, sigs, "")
	assert.Equal(t, msg, tx.GetMsg())
	assert.Equal(t, sigs, tx.GetSignatures())

	feePayer := FeePayer(tx)
	assert.Equal(t, addr, feePayer)
}
