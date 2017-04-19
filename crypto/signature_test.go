package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/ed25519"
	data "github.com/tendermint/go-wire/data"
)

func TestSignAndValidateEd25519(t *testing.T) {

	privKey := GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sigEd := sig.Unwrap().(SignatureEd25519)
	sigEd[7] ^= byte(0x01)
	sig = sigEd.Wrap()

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignAndValidateSecp256k1(t *testing.T) {
	privKey := GenPrivKeySecp256k1()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)

	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sigEd := sig.Unwrap().(SignatureSecp256k1)
	sigEd[3] ^= byte(0x01)
	sig = sigEd.Wrap()

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignatureEncodings(t *testing.T) {
	cases := []struct {
		privKey PrivKey
		sigSize int
		sigType byte
		sigName string
	}{
		{
			privKey: GenPrivKeyEd25519().Wrap(),
			sigSize: ed25519.SignatureSize,
			sigType: TypeEd25519,
			sigName: NameEd25519,
		},
		{
			privKey: GenPrivKeySecp256k1().Wrap(),
			sigSize: 0, // unknown
			sigType: TypeSecp256k1,
			sigName: NameSecp256k1,
		},
	}

	for _, tc := range cases {
		// note we embed them from the beginning....
		pubKey := tc.privKey.PubKey()

		msg := CRandBytes(128)
		sig := tc.privKey.Sign(msg)

		// store as wire
		bin, err := data.ToWire(sig)
		require.Nil(t, err, "%+v", err)
		if tc.sigSize != 0 {
			assert.Equal(t, tc.sigSize+1, len(bin))
		}
		assert.Equal(t, tc.sigType, bin[0])

		// and back
		sig2 := Signature{}
		err = data.FromWire(bin, &sig2)
		require.Nil(t, err, "%+v", err)
		assert.EqualValues(t, sig, sig2)
		assert.True(t, pubKey.VerifyBytes(msg, sig2))

		// store as json
		js, err := data.ToJSON(sig)
		require.Nil(t, err, "%+v", err)
		assert.True(t, strings.Contains(string(js), tc.sigName))

		// and back
		sig3 := Signature{}
		err = data.FromJSON(js, &sig3)
		require.Nil(t, err, "%+v", err)
		assert.EqualValues(t, sig, sig3)
		assert.True(t, pubKey.VerifyBytes(msg, sig3))

		// and make sure we can textify it
		text, err := data.ToText(sig)
		require.Nil(t, err, "%+v", err)
		assert.True(t, strings.HasPrefix(text, tc.sigName))
	}
}

func TestWrapping(t *testing.T) {
	assert := assert.New(t)

	// construct some basic constructs
	msg := CRandBytes(128)
	priv := GenPrivKeyEd25519()
	pub := priv.PubKey()
	sig := priv.Sign(msg)

	// do some wrapping
	pubs := []PubKey{
		PubKey{nil},
		pub.Wrap(),
		pub.Wrap().Wrap().Wrap(),
		PubKey{PubKey{PubKey{pub}}}.Wrap(),
	}
	for _, p := range pubs {
		_, ok := p.PubKeyInner.(PubKey)
		assert.False(ok)
	}

	sigs := []Signature{
		Signature{nil},
		sig.Wrap(),
		sig.Wrap().Wrap().Wrap(),
		Signature{Signature{Signature{sig}}}.Wrap(),
	}
	for _, s := range sigs {
		_, ok := s.SignatureInner.(Signature)
		assert.False(ok)
	}

}
