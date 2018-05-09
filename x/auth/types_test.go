package auth

import (
	"encoding/json"
	"testing"

	"github.com/magiconair/properties/assert"
	crypto "github.com/tendermint/go-crypto"
)

var _ bam.Msg = (*TestMsg)(nil)

// msg type for testing
type TestMsg struct {
	signers []bam.Address
}

func NewTestMsg(addrs ...bam.Address) *TestMsg {
	return &TestMsg{
		signers: addrs,
	}
}

//nolint
func (msg *TestMsg) Type() string { return "TestMsg" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.signers)
	if err != nil {
		panic(err)
	}
	return bz
}
func (msg *TestMsg) ValidateBasic() bam.Error { return nil }
func (msg *TestMsg) GetSigners() []bam.Address {
	return msg.signers
}

func newStdFee() StdFee {
	return NewStdFee(100,
		bam.Coin{"atom", 150},
	)
}

func TestStdTx(t *testing.T) {
	priv := crypto.GenPrivKeyEd25519()
	addr := priv.PubKey().Address()
	msg := NewTestMsg(addr)
	fee := newStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msg, fee, sigs)
	assert.Equal(t, msg, tx.GetMsg())
	assert.Equal(t, sigs, tx.GetSignatures())

	feePayer := FeePayer(tx)
	assert.Equal(t, addr, feePayer)
}
