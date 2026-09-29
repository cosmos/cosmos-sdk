package secp256k1eth_test

import (
	"testing"

	cmtsecp256k1eth "github.com/cometbft/cometbft/crypto/secp256k1eth"
	"github.com/stretchr/testify/require"

	codeccrypto "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1eth"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestSignAndVerify(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	require.Len(t, priv.Bytes(), cmtsecp256k1eth.PrivKeySize)

	pub := priv.PubKey()
	require.NotNil(t, pub)
	require.Equal(t, cmtsecp256k1eth.KeyType, pub.Type())
	require.Len(t, pub.Bytes(), cmtsecp256k1eth.PubKeySize)
	require.Len(t, pub.Address(), 20)

	msg := []byte("hello secp256k1eth")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)
	require.Len(t, sig, cmtsecp256k1eth.SignatureSize)

	require.True(t, pub.VerifySignature(msg, sig))

	// Tamper with the message.
	tampered := append([]byte(nil), msg...)
	tampered[0] ^= 0xff
	require.False(t, pub.VerifySignature(tampered, sig))

	// Truncated signature.
	require.False(t, pub.VerifySignature(msg, sig[:64]))
}

// TestVerifySignatureMalformedKey pins the behavior for keys that bypass
// constructor validation (proto decoding enforces no length or point check):
// verification must return false, not panic. The cometbft implementation never
// parses the key — it byte-compares against the signature's recovered key.
func TestVerifySignatureMalformedKey(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	msg := []byte("msg")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	// Correct length but not a valid curve point (invalid SEC1 prefix).
	bad := make([]byte, cmtsecp256k1eth.PubKeySize)
	bad[0] = 0x05
	badPub := &secp256k1eth.PubKey{Key: bad}
	require.NotPanics(t, func() { require.False(t, badPub.VerifySignature(msg, sig)) })

	// Wrong length entirely.
	shortPub := &secp256k1eth.PubKey{Key: []byte{0x02}}
	require.NotPanics(t, func() { require.False(t, shortPub.VerifySignature(msg, sig)) })

	// Nil key.
	nilPub := &secp256k1eth.PubKey{}
	require.NotPanics(t, func() { require.False(t, nilPub.VerifySignature(msg, sig)) })
}

func TestRoundTripFromBytes(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()

	got, err := secp256k1eth.NewPrivateKeyFromBytes(priv.Bytes())
	require.NoError(t, err)
	require.True(t, priv.Equals(&got))
	require.True(t, priv.PubKey().Equals(got.PubKey()))

	pub, err := secp256k1eth.NewPubKeyFromBytes(priv.PubKey().Bytes())
	require.NoError(t, err)
	require.True(t, priv.PubKey().Equals(&pub))

	_, err = secp256k1eth.NewPrivateKeyFromBytes(make([]byte, cmtsecp256k1eth.PrivKeySize-1))
	require.Error(t, err)
	_, err = secp256k1eth.NewPubKeyFromBytes(make([]byte, cmtsecp256k1eth.PubKeySize))
	require.Error(t, err)
}

// TestAddressMatchesComet pins the SDK address to cometbft's Ethereum address
// derivation: slashing/evidence look validators up by the consensus address
// cometbft reports, so the two must agree byte-for-byte.
func TestAddressMatchesComet(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	cmtPriv := cmtsecp256k1eth.PrivKey(priv.Bytes())
	require.Equal(t, cmtPriv.PubKey().Address().Bytes(), priv.PubKey().Address().Bytes())
}

// TestCmtBridge exercises the FromCmtPubKeyInterface / ToCmtPubKeyInterface path
// which is the integration point used when handing pubkeys to ABCI / staking.
func TestCmtBridge(t *testing.T) {
	priv := secp256k1eth.GenPrivKey()
	sdkPub := priv.PubKey()

	tmPub, err := codeccrypto.ToCmtPubKeyInterface(sdkPub)
	require.NoError(t, err)
	require.Equal(t, cmtsecp256k1eth.KeyType, tmPub.Type())

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

func TestGenPrivKeyFromSecretDeterministic(t *testing.T) {
	secret := []byte("reproducible e2e secret")

	a := secp256k1eth.GenPrivKeyFromSecret(secret)
	b := secp256k1eth.GenPrivKeyFromSecret(secret)
	require.Equal(t, a.Bytes(), b.Bytes())

	// Matches cometbft's derivation for the same secret.
	require.Equal(t, cmtsecp256k1eth.GenPrivKeySecp256k1Eth(secret).Bytes(), a.Bytes())

	c := secp256k1eth.GenPrivKeyFromSecret([]byte("different secret"))
	require.NotEqual(t, a.Bytes(), c.Bytes())
}

// TestConsensusPubKeyTypeAccepted covers the staking admission check: the key's
// Type() must match cometbft's PubKeyTypes entry for secp256k1eth validators to
// be accepted by MsgCreateValidator / MsgRotateConsPubKey.
func TestConsensusPubKeyTypeAccepted(t *testing.T) {
	pub := secp256k1eth.GenPrivKey().PubKey()
	require.NoError(t, types.ValidateConsensusPubKeyType(pub, []string{"ed25519", "secp256k1eth"}))
	require.Error(t, types.ValidateConsensusPubKeyType(pub, []string{"ed25519"}))
}
