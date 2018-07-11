package types

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
)

var invalidstrs = []string{
	"",
	"hello, world!",
	"0xAA",
	"AAA",
	Bech32PrefixAccAddr + "AB0C",
	Bech32PrefixAccPub + "1234",
	Bech32PrefixValAddr + "5678",
	Bech32PrefixValPub + "BBAB",
}

func TestPubKey(t *testing.T) {
	var pub crypto.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		mustbech32accpub := MustBech32ifyAccPub(pub)
		bech32accpub, err := Bech32ifyAccPub(pub)
		require.Nil(t, err)
		require.Equal(t, bech32accpub, mustbech32accpub)

		mustbech32valpub := MustBech32ifyValPub(pub)
		bech32valpub, err := Bech32ifyValPub(pub)
		require.Nil(t, err)
		require.Equal(t, bech32valpub, mustbech32valpub)

		mustaccpub := MustGetAccPubKeyBech32(bech32accpub)
		accpub, err := GetAccPubKeyBech32(bech32accpub)
		require.Nil(t, err)
		require.Equal(t, accpub, mustaccpub)

		mustvalpub := MustGetValPubKeyBech32(bech32valpub)
		valpub, err := GetValPubKeyBech32(bech32valpub)
		require.Nil(t, err)
		require.Equal(t, valpub, mustvalpub)

		require.Equal(t, valpub, accpub)
	}
}

func TestAccAddr(t *testing.T) {
	var pub crypto.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := AccAddress(pub.Address())
		res := AccAddress{}

		bz, err := acc.MarshalJSON()
		require.Nil(t, err)
		err = (&res).UnmarshalJSON(bz)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		bz, err = acc.Marshal()
		require.Nil(t, err)
		err = (&res).Unmarshal(bz)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str := acc.String()
		res, err = AccAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = AccAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidstrs {
		_, err := AccAddressFromHex(str)
		require.NotNil(t, err)

		_, err = AccAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*AccAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}

func TestValAddr(t *testing.T) {
	var pub crypto.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := ValAddress(pub.Address())
		res := ValAddress{}

		bz, err := acc.MarshalJSON()
		require.Nil(t, err)
		err = (&res).UnmarshalJSON(bz)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		bz, err = acc.Marshal()
		require.Nil(t, err)
		err = (&res).Unmarshal(bz)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str := acc.String()
		res, err = ValAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = ValAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidstrs {
		_, err := ValAddressFromHex(str)
		require.NotNil(t, err)

		_, err = ValAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*ValAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}
