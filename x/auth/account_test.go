package auth

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	crypto "github.com/tendermint/go-crypto"
)

func TestBaseAccount(t *testing.T) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := pub.Address()
	someCoins := sdk.Coins{{"atom", 123}, {"eth", 246}}
	seq := int64(7)

	acc := NewBaseAccountWithAddress(addr)

	err := acc.SetPubKey(pub)
	assert.Nil(t, err)
	assert.Equal(t, pub, acc.GetPubKey())

	assert.Equal(t, addr, acc.GetAddress())

	err = acc.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.Equal(t, someCoins, acc.GetCoins())

	err = acc.SetSequence(seq)
	assert.Nil(t, err)
	assert.Equal(t, seq, acc.GetSequence())

	b, err := acc.MarshalJSON()
	assert.Nil(t, err)

	acc2 := new(BaseAccount)
	err = acc2.UnmarshalJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, acc, acc2)

	acc2 = new(BaseAccount)
	err = acc2.UnmarshalJSON(b[:len(b)/2])
	assert.NotNil(t, err)

}
