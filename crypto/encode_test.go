package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	tcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type byter interface {
	Bytes() []byte
}

func checkAminoBinary(t *testing.T, src, dst interface{}, size int) {
	// Marshal to binary bytes.
	bz, err := cdc.MarshalBinaryBare(src)
	require.Nil(t, err, "%+v", err)
	if byterSrc, ok := src.(byter); ok {
		// Make sure this is compatible with current (Bytes()) encoding.
		require.Equal(t, byterSrc.Bytes(), bz, "Amino binary vs Bytes() mismatch")
	}
	// Make sure we have the expected length.
	if size != -1 {
		require.Equal(t, size, len(bz), "Amino binary size mismatch")
	}
	// Unmarshal.
	err = cdc.UnmarshalBinaryBare(bz, dst)
	require.Nil(t, err, "%+v", err)
}

func checkAminoJSON(t *testing.T, src interface{}, dst interface{}, isNil bool) {
	// Marshal to JSON bytes.
	js, err := cdc.MarshalJSON(src)
	require.Nil(t, err, "%+v", err)
	if isNil {
		require.Equal(t, string(js), `null`)
	} else {
		require.Contains(t, string(js), `"type":`)
		require.Contains(t, string(js), `"value":`)
	}
	// Unmarshal.
	err = cdc.UnmarshalJSON(js, dst)
	require.Nil(t, err, "%+v", err)
}

// nolint: vet
func ExamplePrintRegisteredTypes() {
	cdc.PrintTypes(os.Stdout)
	// Output: | Type | Name | Prefix | Length | Notes |
	//| ---- | ---- | ------ | ----- | ------ |
	//| PrivKeyLedgerSecp256k1 | tendermint/PrivKeyLedgerSecp256k1 | 0x10CAB393 | variable |  |
	//| PubKeyEd25519 | tendermint/PubKeyEd25519 | 0x1624DE64 | 0x20 |  |
	//| PubKeySecp256k1 | tendermint/PubKeySecp256k1 | 0xEB5AE987 | 0x21 |  |
	//| PubKeyMultisigThreshold | tendermint/PubKeyMultisigThreshold | 0x22C1F7E2 | variable |  |
	//| PrivKeyEd25519 | tendermint/PrivKeyEd25519 | 0xA3288910 | 0x40 |  |
	//| PrivKeySecp256k1 | tendermint/PrivKeySecp256k1 | 0xE1B0F79B | 0x20 |  |
}

func TestKeyEncodings(t *testing.T) {
	cases := []struct {
		privKey           tcrypto.PrivKey
		privSize, pubSize int // binary sizes with the amino overhead
	}{
		{
			privKey:  ed25519.GenPrivKey(),
			privSize: 69,
			pubSize:  37,
		},
		{
			privKey:  secp256k1.GenPrivKey(),
			privSize: 37,
			pubSize:  38,
		},
	}

	for _, tc := range cases {

		// Check (de/en)codings of PrivKeys.
		var priv2, priv3 tcrypto.PrivKey
		checkAminoBinary(t, tc.privKey, &priv2, tc.privSize)
		require.EqualValues(t, tc.privKey, priv2)
		checkAminoJSON(t, tc.privKey, &priv3, false) // TODO also check Prefix bytes.
		require.EqualValues(t, tc.privKey, priv3)

		// Check (de/en)codings of Signatures.
		var sig1, sig2 []byte
		sig1, err := tc.privKey.Sign([]byte("something"))
		require.NoError(t, err)
		checkAminoBinary(t, sig1, &sig2, -1) // Signature size changes for Secp anyways.
		require.EqualValues(t, sig1, sig2)

		// Check (de/en)codings of PubKeys.
		pubKey := tc.privKey.PubKey()
		var pub2, pub3 tcrypto.PubKey
		checkAminoBinary(t, pubKey, &pub2, tc.pubSize)
		require.EqualValues(t, pubKey, pub2)
		checkAminoJSON(t, pubKey, &pub3, false) // TODO also check Prefix bytes.
		require.EqualValues(t, pubKey, pub3)
	}
}

func TestNilEncodings(t *testing.T) {

	// Check nil Signature.
	var a, b []byte
	checkAminoJSON(t, &a, &b, true)
	require.EqualValues(t, a, b)

	// Check nil PubKey.
	var c, d tcrypto.PubKey
	checkAminoJSON(t, &c, &d, true)
	require.EqualValues(t, c, d)

	// Check nil PrivKey.
	var e, f tcrypto.PrivKey
	checkAminoJSON(t, &e, &f, true)
	require.EqualValues(t, e, f)

}
