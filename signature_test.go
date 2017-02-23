package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/ed25519"
	data "github.com/tendermint/go-data"
)

func TestSignAndValidateEd25519(t *testing.T) {

	privKey := GenPrivKeyEd25519()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sigEd := sig.(SignatureEd25519)
	sigEd[0] ^= byte(0x01)
	sig = Signature(sigEd)

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignAndValidateSecp256k1(t *testing.T) {
	privKey := GenPrivKeySecp256k1()
	pubKey := privKey.PubKey()

	msg := CRandBytes(128)
	sig := privKey.Sign(msg)

	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	sigEd := sig.(SignatureSecp256k1)
	sigEd[0] ^= byte(0x01)
	sig = Signature(sigEd)

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignatureEncodings(t *testing.T) {
	cases := []struct {
		privKey PrivKeyS
		sigSize int
		sigType byte
		sigName string
	}{
		{
			privKey: PrivKeyS{GenPrivKeyEd25519()},
			sigSize: ed25519.SignatureSize,
			sigType: SignatureTypeEd25519,
			sigName: SignatureNameEd25519,
		},
		{
			privKey: PrivKeyS{GenPrivKeySecp256k1()},
			sigSize: 0, // unknown
			sigType: SignatureTypeSecp256k1,
			sigName: SignatureNameSecp256k1,
		},
	}

	for _, tc := range cases {
		// note we embed them from the beginning....
		pubKey := PubKeyS{tc.privKey.PubKey()}

		msg := CRandBytes(128)
		sig := SignatureS{tc.privKey.Sign(msg)}

		// store as wire
		bin, err := data.ToWire(sig)
		require.Nil(t, err, "%+v", err)
		if tc.sigSize != 0 {
			assert.Equal(t, tc.sigSize+1, len(bin))
		}
		assert.Equal(t, tc.sigType, bin[0])

		// and back
		sig2 := SignatureS{}
		err = data.FromWire(bin, &sig2)
		require.Nil(t, err, "%+v", err)
		assert.EqualValues(t, sig, sig2)
		assert.True(t, pubKey.VerifyBytes(msg, sig2))

		// store as json
		js, err := data.ToJSON(sig)
		require.Nil(t, err, "%+v", err)
		assert.True(t, strings.Contains(string(js), tc.sigName))

		// and back
		sig3 := SignatureS{}
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
