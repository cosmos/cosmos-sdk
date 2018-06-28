package crypto

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	tcrypto "github.com/tendermint/tendermint/crypto"
)

type byter interface {
	Bytes() []byte
}

func checkAminoBinary(t *testing.T, src byter, dst interface{}, size int) {
	// Marshal to binary bytes.
	bz, err := cdc.MarshalBinaryBare(src)
	require.Nil(t, err, "%+v", err)
	// Make sure this is compatible with current (Bytes()) encoding.
	assert.Equal(t, src.Bytes(), bz, "Amino binary vs Bytes() mismatch")
	// Make sure we have the expected length.
	if size != -1 {
		assert.Equal(t, size, len(bz), "Amino binary size mismatch")
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
		assert.Equal(t, string(js), `null`)
	} else {
		assert.Contains(t, string(js), `"type":`)
		assert.Contains(t, string(js), `"value":`)
	}
	// Unmarshal.
	err = cdc.UnmarshalJSON(js, dst)
	require.Nil(t, err, "%+v", err)
}

//nolint
func ExamplePrintRegisteredTypes() {
	cdc.PrintTypes(os.Stdout)
	// Output: | Type | Name | Prefix | Length | Notes |
	//| ---- | ---- | ------ | ----- | ------ |
	//| PrivKeyLedgerSecp256k1 | tendermint/PrivKeyLedgerSecp256k1 | 0x10CAB393 | variable |  |
	//| PubKeyEd25519 | tendermint/PubKeyEd25519 | 0x1624DE64 | 0x20 |  |
	//| PubKeySecp256k1 | tendermint/PubKeySecp256k1 | 0xEB5AE987 | 0x21 |  |
	//| PrivKeyEd25519 | tendermint/PrivKeyEd25519 | 0xA3288910 | 0x40 |  |
	//| PrivKeySecp256k1 | tendermint/PrivKeySecp256k1 | 0xE1B0F79B | 0x20 |  |
	//| SignatureEd25519 | tendermint/SignatureEd25519 | 0x2031EA53 | 0x40 |  |
	//| SignatureSecp256k1 | tendermint/SignatureSecp256k1 | 0x7FC4A495 | variable |  |
}

func TestKeyEncodings(t *testing.T) {
	cases := []struct {
		privKey           tcrypto.PrivKey
		privSize, pubSize int // binary sizes
	}{
		{
			privKey:  tcrypto.GenPrivKeyEd25519(),
			privSize: 69,
			pubSize:  37,
		},
		{
			privKey:  tcrypto.GenPrivKeySecp256k1(),
			privSize: 37,
			pubSize:  38,
		},
	}

	for _, tc := range cases {

		// Check (de/en)codings of PrivKeys.
		var priv2, priv3 tcrypto.PrivKey
		checkAminoBinary(t, tc.privKey, &priv2, tc.privSize)
		assert.EqualValues(t, tc.privKey, priv2)
		checkAminoJSON(t, tc.privKey, &priv3, false) // TODO also check Prefix bytes.
		assert.EqualValues(t, tc.privKey, priv3)

		// Check (de/en)codings of Signatures.
		var sig1, sig2, sig3 tcrypto.Signature
		sig1, err := tc.privKey.Sign([]byte("something"))
		assert.NoError(t, err)
		checkAminoBinary(t, sig1, &sig2, -1) // Signature size changes for Secp anyways.
		assert.EqualValues(t, sig1, sig2)
		checkAminoJSON(t, sig1, &sig3, false) // TODO also check Prefix bytes.
		assert.EqualValues(t, sig1, sig3)

		// Check (de/en)codings of PubKeys.
		pubKey := tc.privKey.PubKey()
		var pub2, pub3 tcrypto.PubKey
		checkAminoBinary(t, pubKey, &pub2, tc.pubSize)
		assert.EqualValues(t, pubKey, pub2)
		checkAminoJSON(t, pubKey, &pub3, false) // TODO also check Prefix bytes.
		assert.EqualValues(t, pubKey, pub3)
	}
}

func TestNilEncodings(t *testing.T) {

	// Check nil Signature.
	var a, b tcrypto.Signature
	checkAminoJSON(t, &a, &b, true)
	assert.EqualValues(t, a, b)

	// Check nil PubKey.
	var c, d tcrypto.PubKey
	checkAminoJSON(t, &c, &d, true)
	assert.EqualValues(t, c, d)

	// Check nil PrivKey.
	var e, f tcrypto.PrivKey
	checkAminoJSON(t, &e, &f, true)
	assert.EqualValues(t, e, f)

}
