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
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	fee := newStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msgs, fee, sigs, "")
	assert.Equal(t, msgs, tx.GetMsgs())
	assert.Equal(t, sigs, tx.GetSignatures())

	feePayer := FeePayer(tx)
	assert.Equal(t, addr, feePayer)
}
