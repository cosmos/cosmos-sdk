package cryptoamino

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

type aminoMarshler interface {
	AminoMarshal() ([]byte, error)
	AminoUnmarshal([]byte) error
}

func checkAminoBinary(t *testing.T, src, dst interface{}, size int) {
	// Marshal to binary bytes.
	bz, err := cdc.MarshalBinaryBare(src)
	require.Nil(t, err, "%+v", err)
	if byterSrc, ok := src.(aminoMarshler); ok {
		// Make sure this is compatible with current (Bytes()) encoding.
		aminoBytes, err := byterSrc.AminoMarshal()
		assert.NoError(t, err)
		assert.Equal(t, aminoBytes, bz, "Amino binary vs Bytes() mismatch")
	}
	// Make sure we have the expected length.
	assert.Equal(t, size, len(bz), "Amino binary size mismatch")

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

func TestKeyEncodings(t *testing.T) {
	cases := []struct {
		privKey                    crypto.PrivKey
		privSize, pubSize, sigSize int // binary sizes
	}{
		{
			privKey:  ed25519.GenPrivKey(),
			privSize: 69,
			pubSize:  37,
			sigSize:  65,
		},
		{
			privKey:  sr25519.GenPrivKey(),
			privSize: 37,
			pubSize:  37,
			sigSize:  65,
		},
		{
			privKey:  secp256k1.GenPrivKey(),
			privSize: 37,
			pubSize:  38,
			sigSize:  65,
		},
	}

	for tcIndex, tc := range cases {

		// Check (de/en)codings of PrivKeys.
		var priv2, priv3 crypto.PrivKey
		checkAminoBinary(t, tc.privKey, &priv2, tc.privSize)
		assert.EqualValues(t, tc.privKey, priv2, "tc #%d", tcIndex)
		checkAminoJSON(t, tc.privKey, &priv3, false) // TODO also check Prefix bytes.
		assert.EqualValues(t, tc.privKey, priv3, "tc #%d", tcIndex)

		// Check (de/en)codings of Signatures.
		var sig1, sig2 []byte
		sig1, err := tc.privKey.Sign([]byte("something"))
		assert.NoError(t, err, "tc #%d", tcIndex)
		checkAminoBinary(t, sig1, &sig2, tc.sigSize)
		assert.EqualValues(t, sig1, sig2, "tc #%d", tcIndex)

		// Check (de/en)codings of PubKeys.
		pubKey := tc.privKey.PubKey()
		var pub2, pub3 crypto.PubKey
		checkAminoBinary(t, pubKey, &pub2, tc.pubSize)
		assert.EqualValues(t, pubKey, pub2, "tc #%d", tcIndex)
		checkAminoJSON(t, pubKey, &pub3, false) // TODO also check Prefix bytes.
		assert.EqualValues(t, pubKey, pub3, "tc #%d", tcIndex)
	}
}

func TestNilEncodings(t *testing.T) {

	// Check nil Signature.
	var a, b []byte
	checkAminoJSON(t, &a, &b, true)
	assert.EqualValues(t, a, b)

	// Check nil PubKey.
	var c, d crypto.PubKey
	checkAminoJSON(t, &c, &d, true)
	assert.EqualValues(t, c, d)

	// Check nil PrivKey.
	var e, f crypto.PrivKey
	checkAminoJSON(t, &e, &f, true)
	assert.EqualValues(t, e, f)
}

func TestPubKeyInvalidDataProperReturnsEmpty(t *testing.T) {
	pk, err := PubKeyFromBytes([]byte("foo"))
	require.NotNil(t, err)
	require.Nil(t, pk)
}

func TestPubkeyAminoName(t *testing.T) {
	tests := []struct {
		key   crypto.PubKey
		want  string
		found bool
	}{
		{ed25519.PubKey{}, ed25519.PubKeyAminoName, true},
		{sr25519.PubKey{}, sr25519.PubKeyAminoName, true},
		{secp256k1.PubKey{}, secp256k1.PubKeyAminoName, true},
		{multisig.PubKeyMultisigThreshold{}, multisig.PubKeyMultisigThresholdAminoRoute, true},
	}
	for i, tc := range tests {
		got, found := PubkeyAminoName(cdc, tc.key)
		require.Equal(t, tc.found, found, "not equal on tc %d", i)
		if tc.found {
			require.Equal(t, tc.want, got, "not equal on tc %d", i)
		}
	}
}

var _ crypto.PrivKey = testPriv{}
var _ crypto.PubKey = testPub{}
var testCdc = amino.NewCodec()

type testPriv []byte

func (privkey testPriv) PubKey() crypto.PubKey { return testPub{} }
func (privkey testPriv) Bytes() []byte {
	return testCdc.MustMarshalBinaryBare(privkey)
}
func (privkey testPriv) Sign(msg []byte) ([]byte, error)  { return []byte{}, nil }
func (privkey testPriv) Equals(other crypto.PrivKey) bool { return true }

type testPub []byte

func (key testPub) Address() crypto.Address { return crypto.Address{} }
func (key testPub) Bytes() []byte {
	return testCdc.MustMarshalBinaryBare(key)
}
func (key testPub) VerifyBytes(msg []byte, sig []byte) bool { return true }
func (key testPub) Equals(other crypto.PubKey) bool         { return true }

var (
	privAminoName = "registerTest/Priv"
	pubAminoName  = "registerTest/Pub"
)

func TestRegisterKeyType(t *testing.T) {
	RegisterAmino(testCdc)
	testCdc.RegisterConcrete(testPriv{}, privAminoName, nil)
	testCdc.RegisterConcrete(testPub{}, pubAminoName, nil)

	pub := testPub{0x1}
	priv := testPriv{0x2}

	// Check to make sure key cannot be decoded before registering
	_, err := PrivKeyFromBytes(priv.Bytes())
	require.Error(t, err)
	_, err = PubKeyFromBytes(pub.Bytes())
	require.Error(t, err)

	// Check that name is not registered
	_, found := PubkeyAminoName(testCdc, pub)
	require.False(t, found)

	// Register key types
	RegisterKeyType(testPriv{}, privAminoName)
	RegisterKeyType(testPub{}, pubAminoName)

	// Name should exist after registering
	name, found := PubkeyAminoName(testCdc, pub)
	require.True(t, found)
	require.Equal(t, name, pubAminoName)

	// Decode keys using the encoded bytes from encoding with the other codec
	decodedPriv, err := PrivKeyFromBytes(priv.Bytes())
	require.NoError(t, err)
	require.Equal(t, priv, decodedPriv)

	decodedPub, err := PubKeyFromBytes(pub.Bytes())
	require.NoError(t, err)
	require.Equal(t, pub, decodedPub)

	// Reset module codec after testing
	cdc = amino.NewCodec()
	nameTable = make(map[reflect.Type]string, 3)
	RegisterAmino(cdc)
	nameTable[reflect.TypeOf(ed25519.PubKey{})] = ed25519.PubKeyAminoName
	nameTable[reflect.TypeOf(sr25519.PubKey{})] = sr25519.PubKeyAminoName
	nameTable[reflect.TypeOf(secp256k1.PubKey{})] = secp256k1.PubKeyAminoName
	nameTable[reflect.TypeOf(multisig.PubKeyMultisigThreshold{})] = multisig.PubKeyMultisigThresholdAminoRoute
}
