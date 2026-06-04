package mldsa65_test

import (
	"testing"

	cmtmldsa65 "github.com/cometbft/cometbft/crypto/mldsa65"
	"github.com/stretchr/testify/require"

	codeccrypto "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
)

func TestSignAndVerify(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	require.Len(t, priv.Bytes(), cmtmldsa65.PrivKeySize)

	pub := priv.PubKey()
	require.NotNil(t, pub)
	require.Equal(t, cmtmldsa65.KeyType, pub.Type())
	require.Len(t, pub.Bytes(), cmtmldsa65.PubKeySize)
	require.Len(t, pub.Address(), 20)

	msg := []byte("hello mldsa65")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)
	require.Len(t, sig, cmtmldsa65.SignatureSize)

	require.True(t, pub.VerifySignature(msg, sig))

	// Tamper with the message.
	tampered := append([]byte(nil), msg...)
	tampered[0] ^= 0xff
	require.False(t, pub.VerifySignature(tampered, sig))
}

func TestRoundTripFromBytes(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)

	got, err := mldsa65.NewPrivateKeyFromBytes(priv.Bytes())
	require.NoError(t, err)
	require.True(t, priv.Equals(&got))
	require.True(t, priv.PubKey().Equals(got.PubKey()))
}

// TestCmtBridge exercises the FromCmtPubKeyInterface / ToCmtPubKeyInterface path
// which is the integration point used when handing pubkeys to ABCI / staking.
func TestCmtBridge(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	sdkPub := priv.PubKey()

	tmPub, err := codeccrypto.ToCmtPubKeyInterface(sdkPub)
	require.NoError(t, err)
	require.Equal(t, cmtmldsa65.KeyType, tmPub.Type())

	roundtrip, err := codeccrypto.FromCmtPubKeyInterface(tmPub)
	require.NoError(t, err)
	require.True(t, sdkPub.Equals(roundtrip))

	// Signatures produced via the SDK key verify under the cometbft pubkey
	// after the round-trip.
	msg := []byte("bridge round trip")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)
	require.True(t, tmPub.VerifySignature(msg, sig))
}

func TestGenPrivKeyFromSeedDeterministic(t *testing.T) {
	seed := make([]byte, cmtmldsa65.SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}

	a, err := mldsa65.GenPrivKeyFromSeed(seed)
	require.NoError(t, err)
	b, err := mldsa65.GenPrivKeyFromSeed(seed)
	require.NoError(t, err)

	// Same seed -> identical key bytes.
	require.Equal(t, a.Bytes(), b.Bytes())
	require.True(t, a.PubKey().Equals(b.PubKey()))

	// Different seed -> different key.
	seed[0] ^= 0xff
	c, err := mldsa65.GenPrivKeyFromSeed(seed)
	require.NoError(t, err)
	require.NotEqual(t, a.Bytes(), c.Bytes())
}

func TestGenPrivKeyFromSeedWrongSize(t *testing.T) {
	_, err := mldsa65.GenPrivKeyFromSeed(make([]byte, cmtmldsa65.SeedSize-1))
	require.Error(t, err)
}
