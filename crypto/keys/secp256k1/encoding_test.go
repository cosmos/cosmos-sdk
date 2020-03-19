package secp256k1_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

func TestPrivKeySecp256k1Marshal(t *testing.T) {
	testCases := []struct {
		in  string
		out string // amino compatible
	}{
		{
			"BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF1734",
			"E1B0F79B20BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF1734",
		},
		{
			"0000000000000000000000000000000000000000000000000000000000000000",
			"E1B0F79B200000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	for _, tc := range testCases {
		privkeyBz, err := hex.DecodeString(tc.in)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%X", len(privkeyBz)), "20")

		var privkey secp256k1.PrivKey = privkeyBz

		bz, err := privkey.AminoMarshal()
		require.NoError(t, err)
		require.Equal(t, tc.out, fmt.Sprintf("%X", bz))
		require.Equal(t, tc.in, fmt.Sprintf("%X", privkey.Bytes()))
	}
}

func TestPrivKeySecp256k1Unmarshal(t *testing.T) {
	testCases := []struct {
		in        string // amino compatible encoding
		out       string
		expectErr bool
	}{
		{
			"E1B0F79B20BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF1734",
			"BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF1734",
			false,
		},
		{
			"E1B0F79B200000000000000000000000000000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
			false,
		},
		{
			"01B0F79B20BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF1734",
			"",
			true,
		},
		{
			"E1B0F79B20BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF173400",
			"",
			true,
		},
		{
			"E1B0F79B21BE0AEFA1385C7DA218E04C3569EF0E3DDF06D877DD54262D08B97B7F8ECF1734",
			"",
			true,
		},
	}

	for _, tc := range testCases {
		bz, err := hex.DecodeString(tc.in)
		require.NoError(t, err)

		var privkey secp256k1.PrivKey

		require.Equal(t, tc.expectErr, privkey.AminoUnmarshal(bz) != nil)
		require.Equal(t, tc.out, fmt.Sprintf("%X", privkey[:]))
	}
}

func TestPubKeySecp256k1MarshalBinary(t *testing.T) {
	testCases := []struct {
		in  string
		out string // amino compatible
	}{
		{
			"000000000000000000000000000000000000000000000000000000000000000000",
			"EB5AE98721000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			"0217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D9",
			"EB5AE987210217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D9",
		},
	}

	for _, tc := range testCases {
		pubkeyBz, err := hex.DecodeString(tc.in)
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%X", len(pubkeyBz)), "21")

		var pubkey secp256k1.PubKey = pubkeyBz

		bz, err := pubkey.AminoMarshal()
		require.NoError(t, err)
		require.Equal(t, tc.out, fmt.Sprintf("%X", bz))
		require.Equal(t, tc.in, fmt.Sprintf("%X", pubkey.Bytes()))
	}
}

func TestPubKeySecp256k1UnmarshalBinary(t *testing.T) {
	testCases := []struct {
		in        string // amino compatible encoding
		out       string
		expectErr bool
	}{
		{
			"EB5AE987210217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D9",
			"PubKeySecp256k1{0217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D9}",
			false,
		},
		{
			"EB5AE98721000000000000000000000000000000000000000000000000000000000000000000",
			"PubKeySecp256k1{000000000000000000000000000000000000000000000000000000000000000000}",
			false,
		},
		{
			"0B5AE987210217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D9",
			"PubKeySecp256k1{}",
			true,
		},
		{
			"EB5AE987200217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D9",
			"PubKeySecp256k1{}",
			true,
		},
		{
			"EB5AE987210217C937E7BB8DA08C6564C5BD85C7489ED7D663B7A4A301A6F5D865F0419053D900",
			"PubKeySecp256k1{}",
			true,
		},
	}

	for _, tc := range testCases {
		bz, err := hex.DecodeString(tc.in)
		require.NoError(t, err)

		var pubkey secp256k1.PubKey

		require.Equal(t, tc.expectErr, pubkey.AminoUnmarshal(bz) != nil)
		require.Equal(t, tc.out, pubkey.String())
	}
}
