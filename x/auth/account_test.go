package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.Address) {
	key := crypto.GenPrivKeyEd25519()
	pub := key.PubKey()
	addr := pub.Address()
	return key, pub, addr
}

func TestBaseAddressPubKey(t *testing.T) {
	_, pub1, addr1 := keyPubAddr()
	_, pub2, addr2 := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr1)

	// check the address (set) and pubkey (not set)
	assert.EqualValues(t, addr1, acc.GetAddress())
	assert.EqualValues(t, nil, acc.GetPubKey())

	// can't override address
	err := acc.SetAddress(addr2)
	assert.NotNil(t, err)
	assert.EqualValues(t, addr1, acc.GetAddress())

	// set the pubkey
	err = acc.SetPubKey(pub1)
	assert.Nil(t, err)
	assert.Equal(t, pub1, acc.GetPubKey())

	// can override pubkey
	err = acc.SetPubKey(pub2)
	assert.Nil(t, err)
	assert.Equal(t, pub2, acc.GetPubKey())

	//------------------------------------

	// can set address on empty account
	acc2 := BaseAccount{}
	err = acc2.SetAddress(addr2)
	assert.Nil(t, err)
	assert.EqualValues(t, addr2, acc2.GetAddress())
}

func TestBaseAccountCoins(t *testing.T) {
	_, _, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewCoin("atom", 123), sdk.NewCoin("eth", 246)}

	err := acc.SetCoins(someCoins)
	assert.Nil(t, err)
	assert.Equal(t, someCoins, acc.GetCoins())
}

func TestBaseAccountSequence(t *testing.T) {
	_, _, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	seq := int64(7)

	err := acc.SetSequence(seq)
	assert.Nil(t, err)
	assert.Equal(t, seq, acc.GetSequence())
}

func TestBaseAccountMarshal(t *testing.T) {
	_, pub, addr := keyPubAddr()
	acc := NewBaseAccountWithAddress(addr)

	someCoins := sdk.Coins{sdk.NewCoin("atom", 123), sdk.NewCoin("eth", 246)}
	seq := int64(7)

	// set everything on the account
	err := acc.SetPubKey(pub)
	assert.Nil(t, err)
	err = acc.SetSequence(seq)
	assert.Nil(t, err)
	err = acc.SetCoins(someCoins)
	assert.Nil(t, err)

	// need a codec for marshaling
	codec := wire.NewCodec()
	wire.RegisterCrypto(codec)

	b, err := codec.MarshalBinary(acc)
	assert.Nil(t, err)

	acc2 := BaseAccount{}
	err = codec.UnmarshalBinary(b, &acc2)
	assert.Nil(t, err)
	assert.Equal(t, acc, acc2)

	// error on bad bytes
	acc2 = BaseAccount{}
	err = codec.UnmarshalBinary(b[:len(b)/2], &acc2)
	assert.NotNil(t, err)

}
