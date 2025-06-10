package keys_test

import (
	"testing"

	"github.com/cometbft/cometbft/v2/crypto/bls12381"
	"github.com/cometbft/cometbft/v2/crypto/ed25519"
	"github.com/cometbft/cometbft/v2/crypto/encoding"
	"github.com/cometbft/cometbft/v2/crypto/secp256k1"
	"github.com/cometbft/cometbft/v2/crypto/secp256k1eth"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func TestPubKeyFromTypeAndBytes(t *testing.T) {
	// --- ed25519 ---
	edPriv := ed25519.GenPrivKey()
	edPub := edPriv.PubKey()

	// valid
	pk, err := keys.PubKeyFromTypeAndBytes(edPub.Type(), edPub.Bytes())
	require.NoError(t, err)
	require.Equal(t, edPub.Type(), pk.Type())
	require.Equal(t, edPub.Bytes(), pk.Bytes())
	require.Equal(t, edPub.Address(), pk.Address())
	require.Equal(t, edPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

	// invalid length
	_, err = keys.PubKeyFromTypeAndBytes(edPub.Type(), edPub.Bytes()[:5])
	require.Error(t, err)
	var invLen encoding.ErrInvalidKeyLen
	require.ErrorAs(t, err, &invLen)

	// --- secp256k1 ---
	secpPriv := secp256k1.GenPrivKey()
	secpPub := secpPriv.PubKey()

	pk, err = keys.PubKeyFromTypeAndBytes(secpPub.Type(), secpPub.Bytes())
	require.NoError(t, err)
	require.Equal(t, secpPub.Type(), pk.Type())
	require.Equal(t, secpPub.Bytes(), pk.Bytes())
	require.Equal(t, secpPub.Address(), pk.Address())
	require.Equal(t, secpPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

	_, err = keys.PubKeyFromTypeAndBytes(secpPub.Type(), secpPub.Bytes()[:5])
	require.Error(t, err)
	require.ErrorAs(t, err, &invLen)

	// --- secp256k1eth (optional) ---
	if secp256k1eth.Enabled {
		ethPriv := secp256k1eth.GenPrivKey()
		ethPub := ethPriv.PubKey()

		pk, err = keys.PubKeyFromTypeAndBytes(ethPub.Type(), ethPub.Bytes())
		require.NoError(t, err)
		require.Equal(t, ethPub.Type(), pk.Type())
		require.Equal(t, ethPub.Bytes(), pk.Bytes())
		require.Equal(t, ethPub.Address(), pk.Address())
		require.Equal(t, ethPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

		_, err = keys.PubKeyFromTypeAndBytes(ethPub.Type(), ethPub.Bytes()[:5])
		require.Error(t, err)
		require.ErrorAs(t, err, &invLen)
	} else {
		// should error if type known but not enabled
		_, err := keys.PubKeyFromTypeAndBytes(secp256k1eth.KeyType, []byte{})
		require.Error(t, err)
	}

	// --- bls12381 (optional) ---
	if bls12381.Enabled {
		blsPriv, err := bls12381.GenPrivKey()
		require.NoError(t, err)
		blsPub := blsPriv.PubKey()

		pk, err = keys.PubKeyFromTypeAndBytes(blsPub.Type(), blsPub.Bytes())
		require.NoError(t, err)
		require.Equal(t, blsPub.Type(), pk.Type())
		require.Equal(t, blsPub.Bytes(), pk.Bytes())
		require.Equal(t, blsPub.Address(), pk.Address())
		require.Equal(t, blsPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

		_, err = keys.PubKeyFromTypeAndBytes(blsPub.Type(), blsPub.Bytes()[:5])
		require.Error(t, err)
		require.ErrorAs(t, err, &invLen)
	} else {
		_, err := keys.PubKeyFromTypeAndBytes(bls12381.KeyType, []byte{})
		require.Error(t, err)
	}

	// --- unsupported type ---
	_, err = keys.PubKeyFromTypeAndBytes("not-a-key", []byte{1, 2, 3})
	require.Error(t, err)
	var unsup encoding.ErrUnsupportedKey
	require.ErrorAs(t, err, &unsup)
}
