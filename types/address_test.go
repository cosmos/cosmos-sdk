package types_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/types"
)

var invalidStrs = []string{
	"hello, world!",
	"0xAA",
	"AAA",
	types.Bech32PrefixAccAddr + "AB0C",
	types.Bech32PrefixAccPub + "1234",
	types.Bech32PrefixValAddr + "5678",
	types.Bech32PrefixValPub + "BBAB",
	types.Bech32PrefixConsAddr + "FF04",
	types.Bech32PrefixConsPub + "6789",
}

func testMarshal(t *testing.T, original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	require.Nil(t, err)
	err = unmarshal(bz)
	require.Nil(t, err)
	require.Equal(t, original, res)
}

func TestEmptyAddresses(t *testing.T) {
	t.Parallel()
	require.Equal(t, (types.AccAddress{}).String(), "")
	require.Equal(t, (types.ValAddress{}).String(), "")
	require.Equal(t, (types.ConsAddress{}).String(), "")

	accAddr, err := types.AccAddressFromBech32("")
	require.True(t, accAddr.Empty())
	require.Nil(t, err)

	valAddr, err := types.ValAddressFromBech32("")
	require.True(t, valAddr.Empty())
	require.Nil(t, err)

	consAddr, err := types.ConsAddressFromBech32("")
	require.True(t, consAddr.Empty())
	require.Nil(t, err)
}

func TestRandBech32PubkeyConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		mustBech32AccPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeAccPub, pub)
		bech32AccPub, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeAccPub, pub)
		require.Nil(t, err)
		require.Equal(t, bech32AccPub, mustBech32AccPub)

		mustBech32ValPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeValPub, pub)
		bech32ValPub, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeValPub, pub)
		require.Nil(t, err)
		require.Equal(t, bech32ValPub, mustBech32ValPub)

		mustBech32ConsPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pub)
		bech32ConsPub, err := types.Bech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pub)
		require.Nil(t, err)
		require.Equal(t, bech32ConsPub, mustBech32ConsPub)

		mustAccPub := types.MustGetPubKeyFromBech32(types.Bech32PubKeyTypeAccPub, bech32AccPub)
		accPub, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeAccPub, bech32AccPub)
		require.Nil(t, err)
		require.Equal(t, accPub, mustAccPub)

		mustValPub := types.MustGetPubKeyFromBech32(types.Bech32PubKeyTypeValPub, bech32ValPub)
		valPub, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeValPub, bech32ValPub)
		require.Nil(t, err)
		require.Equal(t, valPub, mustValPub)

		mustConsPub := types.MustGetPubKeyFromBech32(types.Bech32PubKeyTypeConsPub, bech32ConsPub)
		consPub, err := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeConsPub, bech32ConsPub)
		require.Nil(t, err)
		require.Equal(t, consPub, mustConsPub)

		require.Equal(t, valPub, accPub)
		require.Equal(t, valPub, consPub)
	}
}

func TestYAMLMarshalers(t *testing.T) {
	addr := secp256k1.GenPrivKey().PubKey().Address()

	acc := types.AccAddress(addr)
	val := types.ValAddress(addr)
	cons := types.ConsAddress(addr)

	got, _ := yaml.Marshal(&acc)
	require.Equal(t, acc.String()+"\n", string(got))

	got, _ = yaml.Marshal(&val)
	require.Equal(t, val.String()+"\n", string(got))

	got, _ = yaml.Marshal(&cons)
	require.Equal(t, cons.String()+"\n", string(got))
}

func TestRandBech32AccAddrConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		acc := types.AccAddress(pub.Address())
		res := types.AccAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.AccAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.AccAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.AccAddressFromHex(str)
		require.NotNil(t, err)

		_, err = types.AccAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*types.AccAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}

	_, err := types.AccAddressFromHex("")
	require.Equal(t, "decoding Bech32 address failed: must provide an address", err.Error())
}

func TestValAddr(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := types.ValAddress(pub.Address())
		res := types.ValAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.ValAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.ValAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

	}

	for _, str := range invalidStrs {
		_, err := types.ValAddressFromHex(str)
		require.NotNil(t, err)

		_, err = types.ValAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*types.ValAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}

	// test empty string
	_, err := types.ValAddressFromHex("")
	require.Equal(t, "decoding Bech32 address failed: must provide an address", err.Error())
}

func TestConsAddress(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := types.ConsAddress(pub.Address())
		res := types.ConsAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.ConsAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.ConsAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.ConsAddressFromHex(str)
		require.NotNil(t, err)

		_, err = types.ConsAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*types.ConsAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}

	// test empty string
	_, err := types.ConsAddressFromHex("")
	require.Equal(t, "decoding Bech32 address failed: must provide an address", err.Error())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestConfiguredPrefix(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	for length := 1; length < 10; length++ {
		for times := 1; times < 20; times++ {
			rand.Read(pub[:])
			// Test if randomly generated prefix of a given length works
			prefix := RandString(length)

			// Assuming that GetConfig is not sealed.
			config := types.GetConfig()
			config.SetBech32PrefixForAccount(
				prefix+types.PrefixAccount,
				prefix+types.PrefixPublic)

			acc := types.AccAddress(pub.Address())
			require.True(t, strings.HasPrefix(
				acc.String(),
				prefix+types.PrefixAccount), acc.String())

			bech32Pub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeAccPub, pub)
			require.True(t, strings.HasPrefix(
				bech32Pub,
				prefix+types.PrefixPublic))

			config.SetBech32PrefixForValidator(
				prefix+types.PrefixValidator+types.PrefixAddress,
				prefix+types.PrefixValidator+types.PrefixPublic)

			val := types.ValAddress(pub.Address())
			require.True(t, strings.HasPrefix(
				val.String(),
				prefix+types.PrefixValidator+types.PrefixAddress))

			bech32ValPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeValPub, pub)
			require.True(t, strings.HasPrefix(
				bech32ValPub,
				prefix+types.PrefixValidator+types.PrefixPublic))

			config.SetBech32PrefixForConsensusNode(
				prefix+types.PrefixConsensus+types.PrefixAddress,
				prefix+types.PrefixConsensus+types.PrefixPublic)

			cons := types.ConsAddress(pub.Address())
			require.True(t, strings.HasPrefix(
				cons.String(),
				prefix+types.PrefixConsensus+types.PrefixAddress))

			bech32ConsPub := types.MustBech32ifyPubKey(types.Bech32PubKeyTypeConsPub, pub)
			require.True(t, strings.HasPrefix(
				bech32ConsPub,
				prefix+types.PrefixConsensus+types.PrefixPublic))
		}
	}
}

func TestAddressInterface(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	rand.Read(pub[:])

	addrs := []types.Address{
		types.ConsAddress(pub.Address()),
		types.ValAddress(pub.Address()),
		types.AccAddress(pub.Address()),
	}

	for _, addr := range addrs {
		switch addr := addr.(type) {
		case types.AccAddress:
			_, err := types.AccAddressFromBech32(addr.String())
			require.Nil(t, err)
		case types.ValAddress:
			_, err := types.ValAddressFromBech32(addr.String())
			require.Nil(t, err)
		case types.ConsAddress:
			_, err := types.ConsAddressFromBech32(addr.String())
			require.Nil(t, err)
		default:
			t.Fail()
		}
	}

}

func TestCustomAddressVerifier(t *testing.T) {
	// Create a 10 byte address
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	accBech := types.AccAddress(addr).String()
	valBech := types.ValAddress(addr).String()
	consBech := types.ConsAddress(addr).String()
	// Verifiy that the default logic rejects this 10 byte address
	err := types.VerifyAddressFormat(addr)
	require.NotNil(t, err)
	_, err = types.AccAddressFromBech32(accBech)
	require.NotNil(t, err)
	_, err = types.ValAddressFromBech32(valBech)
	require.NotNil(t, err)
	_, err = types.ConsAddressFromBech32(consBech)
	require.NotNil(t, err)

	// Set a custom address verifier that accepts 10 or 20 byte addresses
	types.GetConfig().SetAddressVerifier(func(bz []byte) error {
		n := len(bz)
		if n == 10 || n == types.AddrLen {
			return nil
		}
		return fmt.Errorf("incorrect address length %d", n)
	})

	// Verifiy that the custom logic accepts this 10 byte address
	err = types.VerifyAddressFormat(addr)
	require.Nil(t, err)
	_, err = types.AccAddressFromBech32(accBech)
	require.Nil(t, err)
	_, err = types.ValAddressFromBech32(valBech)
	require.Nil(t, err)
	_, err = types.ConsAddressFromBech32(consBech)
	require.Nil(t, err)
}

func TestBech32ifyAddressBytes(t *testing.T) {
	addr10byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	type args struct {
		prefix string
		bs     []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty address", args{"prefixA", []byte{}}, "", false},
		{"empty prefix", args{"", addr20byte}, "", true},
		{"10-byte address", args{"prefixA", addr10byte}, "prefixA1qqqsyqcyq5rqwzqfwvmuzx", false},
		{"10-byte address", args{"prefixB", addr10byte}, "prefixB1qqqsyqcyq5rqwzqf4xftmx", false},
		{"20-byte address", args{"prefixA", addr20byte}, "prefixA1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn6j4npq", false},
		{"20-byte address", args{"prefixB", addr20byte}, "prefixB1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn8e9wka", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.Bech32ifyAddressBytes(tt.args.prefix, tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bech32ifyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMustBech32ifyAddressBytes(t *testing.T) {
	addr10byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	type args struct {
		prefix string
		bs     []byte
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantPanic bool
	}{
		{"empty address", args{"prefixA", []byte{}}, "", false},
		{"empty prefix", args{"", addr20byte}, "", true},
		{"10-byte address", args{"prefixA", addr10byte}, "prefixA1qqqsyqcyq5rqwzqfwvmuzx", false},
		{"10-byte address", args{"prefixB", addr10byte}, "prefixB1qqqsyqcyq5rqwzqf4xftmx", false},
		{"20-byte address", args{"prefixA", addr20byte}, "prefixA1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn6j4npq", false},
		{"20-byte address", args{"prefixB", addr20byte}, "prefixB1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn8e9wka", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { types.MustBech32ifyAddressBytes(tt.args.prefix, tt.args.bs) })
				return
			}
			require.Equal(t, tt.want, types.MustBech32ifyAddressBytes(tt.args.prefix, tt.args.bs))
		})
	}
}

func TestAddressTypesEquals(t *testing.T) {
	t.Parallel()
	addr1 := secp256k1.GenPrivKey().PubKey().Address()
	accAddr1 := types.AccAddress(addr1)
	consAddr1 := types.ConsAddress(addr1)
	valAddr1 := types.ValAddress(addr1)

	addr2 := secp256k1.GenPrivKey().PubKey().Address()
	accAddr2 := types.AccAddress(addr2)
	consAddr2 := types.ConsAddress(addr2)
	valAddr2 := types.ValAddress(addr2)

	// equality
	require.True(t, accAddr1.Equals(accAddr1))
	require.True(t, consAddr1.Equals(consAddr1))
	require.True(t, valAddr1.Equals(valAddr1))

	// emptiness
	require.True(t, types.AccAddress{}.Equals(types.AccAddress{}))
	require.True(t, types.AccAddress{}.Equals(types.AccAddress(nil)))
	require.True(t, types.AccAddress(nil).Equals(types.AccAddress{}))
	require.True(t, types.AccAddress(nil).Equals(types.AccAddress(nil)))

	require.True(t, types.ConsAddress{}.Equals(types.ConsAddress{}))
	require.True(t, types.ConsAddress{}.Equals(types.ConsAddress(nil)))
	require.True(t, types.ConsAddress(nil).Equals(types.ConsAddress{}))
	require.True(t, types.ConsAddress(nil).Equals(types.ConsAddress(nil)))

	require.True(t, types.ValAddress{}.Equals(types.ValAddress{}))
	require.True(t, types.ValAddress{}.Equals(types.ValAddress(nil)))
	require.True(t, types.ValAddress(nil).Equals(types.ValAddress{}))
	require.True(t, types.ValAddress(nil).Equals(types.ValAddress(nil)))

	require.False(t, accAddr1.Equals(accAddr2))
	require.Equal(t, accAddr1.Equals(accAddr2), accAddr2.Equals(accAddr1))
	require.False(t, consAddr1.Equals(consAddr2))
	require.Equal(t, consAddr1.Equals(consAddr2), consAddr2.Equals(consAddr1))
	require.False(t, valAddr1.Equals(valAddr2))
	require.Equal(t, valAddr1.Equals(valAddr2), valAddr2.Equals(valAddr1))
}

func TestNilAddressTypesEmpty(t *testing.T) {
	t.Parallel()
	require.True(t, types.AccAddress(nil).Empty())
	require.True(t, types.ConsAddress(nil).Empty())
	require.True(t, types.ValAddress(nil).Empty())
}

func TestGetConsAddress(t *testing.T) {
	t.Parallel()
	pk := secp256k1.GenPrivKey().PubKey()
	require.NotEqual(t, types.GetConsAddress(pk), pk.Address())
	require.True(t, bytes.Equal(types.GetConsAddress(pk).Bytes(), pk.Address().Bytes()))
	require.Panics(t, func() { types.GetConsAddress(crypto.PubKey(nil)) })
}

func TestGetFromBech32(t *testing.T) {
	_, err := types.GetFromBech32("", "prefix")
	require.Error(t, err)
	require.Equal(t, "decoding Bech32 address failed: must provide an address", err.Error())
	_, err = types.GetFromBech32("cosmos1qqqsyqcyq5rqwzqfys8f67", "x")
	require.Error(t, err)
	require.Equal(t, "invalid Bech32 prefix; expected x, got cosmos", err.Error())
}
