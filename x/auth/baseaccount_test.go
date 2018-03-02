package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

func TestBaseAccount(t *testing.T) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := pub.Address()
	someCoins := sdk.Coins{{"atom", 123}, {"eth", 246}}
	seq := int64(7)

	acc := NewBaseAccountWithAddress(addr)

	// need a codec for marshaling
	codec := wire.NewCodec()
	wire.RegisterCrypto(codec)

	err := acc.SetPubKey(pub)
	assert.Nil(t, err)
	assert.Equal(t, pub, acc.GetPubKey())

	assert.EqualValues(t, addr, acc.GetAddress())

	err = acc.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.Equal(t, someCoins, acc.GetCoins())

	err = acc.SetSequence(seq)
	assert.Nil(t, err)
	assert.Equal(t, seq, acc.GetSequence())

	b, err := codec.MarshalBinary(acc)
	assert.Nil(t, err)

	var acc2 BaseAccount
	err = codec.UnmarshalBinary(b, &acc2)
	assert.Nil(t, err)
	assert.Equal(t, acc, acc2)

	acc2 = BaseAccount{}
	err = codec.UnmarshalBinary(b[:len(b)/2], &acc2)
	assert.NotNil(t, err)
}
