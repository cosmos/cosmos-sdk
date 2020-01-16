package types_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var invalidStrs = []string{
	"hello, world!",
	"0xAA",
	"AAA",
	sdk.Bech32PrefixAccAddr + "AB0C",
	sdk.Bech32PrefixAccPub + "1234",
	sdk.Bech32PrefixValAddr + "5678",
	sdk.Bech32PrefixValPub + "BBAB",
	sdk.Bech32PrefixConsAddr + "FF04",
	sdk.Bech32PrefixConsPub + "6789",
}

func testMarshal(t *testing.T, original interface{}, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	require.Nil(t, err)
	err = unmarshal(bz)
	require.Nil(t, err)
	require.Equal(t, original, res)
}

func TestEmptyAddresses(t *testing.T) {
	require.Equal(t, (sdk.AccAddress{}).String(), "")
	require.Equal(t, (sdk.ValAddress{}).String(), "")
	require.Equal(t, (sdk.ConsAddress{}).String(), "")

	config := sdk.NewDefaultConfig()
	accAddr, err := sdk.AccAddressFromBech32(config, "")
	require.True(t, accAddr.Empty())
	require.Nil(t, err)

	valAddr, err := sdk.ValAddressFromBech32("")
	require.True(t, valAddr.Empty())
	require.Nil(t, err)

	consAddr, err := sdk.ConsAddressFromBech32(config, "")
	require.True(t, consAddr.Empty())
	require.Nil(t, err)
}

func TestRandBech32PubkeyConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	config := sdk.NewDefaultConfig()

	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		mustBech32AccPub := sdk.MustBech32ifyPubKey(config, sdk.Bech32PubKeyTypeAccPub, pub)
		bech32AccPub, err := sdk.Bech32ifyPubKey(config, sdk.Bech32PubKeyTypeAccPub, pub)
		require.Nil(t, err)
		require.Equal(t, bech32AccPub, mustBech32AccPub)

		mustBech32ValPub := sdk.MustBech32ifyPubKey(config, sdk.Bech32PubKeyTypeValPub, pub)
		bech32ValPub, err := sdk.Bech32ifyPubKey(config, sdk.Bech32PubKeyTypeValPub, pub)
		require.Nil(t, err)
		require.Equal(t, bech32ValPub, mustBech32ValPub)

		mustBech32ConsPub := sdk.MustBech32ifyPubKey(config, sdk.Bech32PubKeyTypeConsPub, pub)
		bech32ConsPub, err := sdk.Bech32ifyPubKey(config, sdk.Bech32PubKeyTypeConsPub, pub)
		require.Nil(t, err)
		require.Equal(t, bech32ConsPub, mustBech32ConsPub)

		mustAccPub := sdk.MustGetPubKeyFromBech32(config, sdk.Bech32PubKeyTypeAccPub, bech32AccPub)
		accPub, err := sdk.GetPubKeyFromBech32(config, sdk.Bech32PubKeyTypeAccPub, bech32AccPub)
		require.Nil(t, err)
		require.Equal(t, accPub, mustAccPub)

		mustValPub := sdk.MustGetPubKeyFromBech32(config, sdk.Bech32PubKeyTypeValPub, bech32ValPub)
		valPub, err := sdk.GetPubKeyFromBech32(config, sdk.Bech32PubKeyTypeValPub, bech32ValPub)
		require.Nil(t, err)
		require.Equal(t, valPub, mustValPub)

		mustConsPub := sdk.MustGetPubKeyFromBech32(config, sdk.Bech32PubKeyTypeConsPub, bech32ConsPub)
		consPub, err := sdk.GetPubKeyFromBech32(config, sdk.Bech32PubKeyTypeConsPub, bech32ConsPub)
		require.Nil(t, err)
		require.Equal(t, consPub, mustConsPub)

		require.Equal(t, valPub, accPub)
		require.Equal(t, valPub, consPub)
	}
}

func TestYAMLMarshalers(t *testing.T) {
	addr := secp256k1.GenPrivKey().PubKey().Address()

	acc := sdk.AccAddress(addr)
	val := sdk.ValAddress(addr)
	cons := sdk.ConsAddress(addr)

	got, _ := yaml.Marshal(&acc)
	require.Equal(t, acc.String()+"\n", string(got))

	got, _ = yaml.Marshal(&val)
	require.Equal(t, val.String()+"\n", string(got))

	got, _ = yaml.Marshal(&cons)
	require.Equal(t, cons.String()+"\n", string(got))
}

func TestRandBech32AccAddrConsistency(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	config := sdk.NewDefaultConfig()
	for i := 0; i < 1000; i++ {
		rand.Read(pub[:])

		acc := sdk.AccAddress(pub.Address())
		res := sdk.AccAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := sdk.AccAddressFromBech32(config, str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = sdk.AccAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := sdk.AccAddressFromHex(str)
		require.NotNil(t, err)

		_, err = sdk.AccAddressFromBech32(config, str)
		require.NotNil(t, err)

		err = (*sdk.AccAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}

func TestValAddr(t *testing.T) {
	var pub ed25519.PubKeyEd25519

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := sdk.ValAddress(pub.Address())
		res := sdk.ValAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := sdk.ValAddressFromBech32(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = sdk.ValAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := sdk.ValAddressFromHex(str)
		require.NotNil(t, err)

		_, err = sdk.ValAddressFromBech32(str)
		require.NotNil(t, err)

		err = (*sdk.ValAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
}

func TestConsAddress(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	config := sdk.NewDefaultConfig()

	for i := 0; i < 20; i++ {
		rand.Read(pub[:])

		acc := sdk.ConsAddress(pub.Address())
		res := sdk.ConsAddress{}

		testMarshal(t, &acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		testMarshal(t, &acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := sdk.ConsAddressFromBech32(config, str)
		require.Nil(t, err)
		require.Equal(t, acc, res)

		str = hex.EncodeToString(acc)
		res, err = sdk.ConsAddressFromHex(str)
		require.Nil(t, err)
		require.Equal(t, acc, res)
	}

	for _, str := range invalidStrs {
		_, err := sdk.ConsAddressFromHex(str)
		require.NotNil(t, err)

		_, err = sdk.ConsAddressFromBech32(config, str)
		require.NotNil(t, err)

		err = (*sdk.ConsAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		require.NotNil(t, err)
	}
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
			config := sdk.GetConfig()
			config.SetBech32PrefixForAccount(
				prefix+sdk.PrefixAccount,
				prefix+sdk.PrefixPublic)

			acc := sdk.AccAddress(pub.Address())
			require.True(t, strings.HasPrefix(
				acc.String(),
				prefix+sdk.PrefixAccount), acc.String())

			bech32Pub := sdk.MustBech32ifyPubKey(config, sdk.Bech32PubKeyTypeAccPub, pub)
			require.True(t, strings.HasPrefix(
				bech32Pub,
				prefix+sdk.PrefixPublic))

			config.SetBech32PrefixForValidator(
				prefix+sdk.PrefixValidator+sdk.PrefixAddress,
				prefix+sdk.PrefixValidator+sdk.PrefixPublic)

			val := sdk.ValAddress(pub.Address())
			require.True(t, strings.HasPrefix(
				val.String(),
				prefix+sdk.PrefixValidator+sdk.PrefixAddress))

			bech32ValPub := sdk.MustBech32ifyPubKey(config, sdk.Bech32PubKeyTypeValPub, pub)
			require.True(t, strings.HasPrefix(
				bech32ValPub,
				prefix+sdk.PrefixValidator+sdk.PrefixPublic))

			config.SetBech32PrefixForConsensusNode(
				prefix+sdk.PrefixConsensus+sdk.PrefixAddress,
				prefix+sdk.PrefixConsensus+sdk.PrefixPublic)

			cons := sdk.ConsAddress(pub.Address())
			require.True(t, strings.HasPrefix(
				cons.String(),
				prefix+sdk.PrefixConsensus+sdk.PrefixAddress))

			bech32ConsPub := sdk.MustBech32ifyPubKey(config, sdk.Bech32PubKeyTypeConsPub, pub)
			require.True(t, strings.HasPrefix(
				bech32ConsPub,
				prefix+sdk.PrefixConsensus+sdk.PrefixPublic))
		}
	}
}

func TestAddressInterface(t *testing.T) {
	var pub ed25519.PubKeyEd25519
	rand.Read(pub[:])
	config := sdk.NewDefaultConfig()
	oldSDKConfig := *sdk.SDKConfig
	sdk.SDKConfig = config
	defer func() { sdk.SDKConfig = &oldSDKConfig }()

	addrs := []sdk.Address{
		sdk.ConsAddress(pub.Address()),
		sdk.ValAddress(pub.Address()),
		sdk.AccAddress(pub.Address()),
	}

	for _, addr := range addrs {
		switch addr := addr.(type) {
		case sdk.AccAddress:
			_, err := sdk.AccAddressFromBech32(config, addr.String())
			require.Nil(t, err)
		case sdk.ValAddress:
			_, err := sdk.ValAddressFromBech32(addr.String())
			require.Nil(t, err)
		case sdk.ConsAddress:
			_, err := sdk.ConsAddressFromBech32(config, addr.String())
			require.Nil(t, err)
		default:
			t.Fail()
		}
	}

}

func TestCustomAddressVerifier(t *testing.T) {
	// Create a 10 byte address
	addr := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	accBech := sdk.AccAddress(addr).String()
	valBech := sdk.ValAddress(addr).String()
	consBech := sdk.ConsAddress(addr).String()
	config := sdk.NewDefaultConfig()
	oldSDKConfig := *sdk.SDKConfig
	sdk.SDKConfig = config
	defer func() { sdk.SDKConfig = &oldSDKConfig }()
	// Verifiy that the default logic rejects this 10 byte address
	err := sdk.VerifyAddressFormat(addr)
	require.NotNil(t, err)
	_, err = sdk.AccAddressFromBech32(config, accBech)
	require.NotNil(t, err)
	_, err = sdk.ValAddressFromBech32(valBech)
	require.NotNil(t, err)
	_, err = sdk.ConsAddressFromBech32(config, consBech)
	require.NotNil(t, err)

	// Set a custom address verifier that accepts 10 or 20 byte addresses
	// TODO: singleton here
	sdk.SDKConfig.SetAddressVerifier(func(bz []byte) error {
		n := len(bz)
		if n == 10 || n == sdk.AddrLen {
			return nil
		}
		return fmt.Errorf("incorrect address length %d", n)
	})

	// Verifiy that the custom logic accepts this 10 byte address
	err = sdk.VerifyAddressFormat(addr)
	require.Nil(t, err)
	_, err = sdk.AccAddressFromBech32(config, accBech)
	require.Nil(t, err)
	_, err = sdk.ValAddressFromBech32(valBech)
	require.Nil(t, err)
	_, err = sdk.ConsAddressFromBech32(config, consBech)
	require.Nil(t, err)
}
