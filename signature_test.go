package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-wire"
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
	sigEd[7] ^= byte(0x01)
	sig = sigEd

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
	sigEd[3] ^= byte(0x01)
	sig = sigEd

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

func TestSignatureEncodings(t *testing.T) {
	cases := []struct {
		privKey   PrivKey
		sigSize   int
		sigPrefix wire.PrefixBytes
	}{
		{
			privKey:   GenPrivKeyEd25519(),
			sigSize:   ed25519.SignatureSize,
			sigPrefix: [4]byte{0xe4, 0x51, 0x7b, 0xa3},
		},
		{
			privKey:   GenPrivKeySecp256k1(),
			sigSize:   0, // unknown
			sigPrefix: [4]byte{0x37, 0xb9, 0x21, 0x3e},
		},
	}

	for _, tc := range cases {
		// note we embed them from the beginning....
		pubKey := tc.privKey.PubKey()

		msg := CRandBytes(128)
		sig := tc.privKey.Sign(msg)

		// store as wire
		bin, err := cdc.MarshalBinary(sig)
		require.Nil(t, err, "%+v", err)
		if tc.sigSize != 0 {
			assert.Equal(t, tc.sigSize+4, len(bin))
		}
		assert.Equal(t, tc.sigPrefix[:], bin[0:4])

		// and back
		sig2 := Signature(nil)
		err = cdc.UnmarshalBinary(bin, &sig2)
		require.Nil(t, err, "%+v", err)
		assert.EqualValues(t, sig, sig2)
		assert.True(t, pubKey.VerifyBytes(msg, sig2))

		/*
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
		*/
	}
}
