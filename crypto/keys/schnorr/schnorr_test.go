package schnorr_test

import (
	"github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/schnorr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSignAndValidateSchnorr(t *testing.T) {
	privKey := schnorr.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(1000)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

	// Mutate the signature
	sig[1] ^= byte(0x01)
	assert.False(t, pubKey.VerifySignature(msg, sig))
}

func TestPubKeyEquals(t *testing.T) {
	schnorrKey := *schnorr.GenPrivKey().PubKey()

	testCases := []struct {
		msg      string
		pubKey   schnorr.PubKey
		other    schnorr.PubKey
		expectEq bool
	}{
		{
			"different bytes",
			schnorrKey,
			*schnorr.GenPrivKey().PubKey(),
			false,
		},
		{
			"equals",
			schnorrKey,
			schnorr.PubKey{
				Key: schnorrKey.Key,
			},
			true,
		},
		//{
		//	"different types",
		//	schnorrKey,
		//	secp256k1.GenPrivKey().PubKey(),
		//	false,
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.pubKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}
