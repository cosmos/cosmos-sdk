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

func TestPubKeyFromCometTypeAndBytes(t *testing.T) {
	// --- ed25519 ---
	edPriv := ed25519.GenPrivKey()
	edPub := edPriv.PubKey()

	// valid
	pk, err := keys.PubKeyFromCometTypeAndBytes(edPub.Type(), edPub.Bytes())
	require.NoError(t, err)
	require.Equal(t, edPub.Type(), pk.Type())
	require.Equal(t, edPub.Bytes(), pk.Bytes())
	require.Equal(t, edPub.Address(), pk.Address())
	require.Equal(t, edPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

	// invalid length
	_, err = keys.PubKeyFromCometTypeAndBytes(edPub.Type(), edPub.Bytes()[:5])
	require.Error(t, err)
	var invLen encoding.ErrInvalidKeyLen
	require.ErrorAs(t, err, &invLen)

	// --- secp256k1 ---
	secpPriv := secp256k1.GenPrivKey()
	secpPub := secpPriv.PubKey()

	pk, err = keys.PubKeyFromCometTypeAndBytes(secpPub.Type(), secpPub.Bytes())
	require.NoError(t, err)
	require.Equal(t, secpPub.Type(), pk.Type())
	require.Equal(t, secpPub.Bytes(), pk.Bytes())
	require.Equal(t, secpPub.Address(), pk.Address())
	require.Equal(t, secpPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

	_, err = keys.PubKeyFromCometTypeAndBytes(secpPub.Type(), secpPub.Bytes()[:5])
	require.Error(t, err)
	require.ErrorAs(t, err, &invLen)

	// --- secp256k1eth (optional) ---
	if secp256k1eth.Enabled {
		ethPriv := secp256k1eth.GenPrivKey()
		ethPub := ethPriv.PubKey()

		pk, err = keys.PubKeyFromCometTypeAndBytes(ethPub.Type(), ethPub.Bytes())
		require.NoError(t, err)
		require.Equal(t, ethPub.Type(), pk.Type())
		require.Equal(t, ethPub.Bytes(), pk.Bytes())
		require.Equal(t, ethPub.Address(), pk.Address())
		require.Equal(t, ethPub.VerifySignature([]byte("msg"), []byte("sig")), pk.VerifySignature([]byte("msg"), []byte("sig")))

		_, err = keys.PubKeyFromCometTypeAndBytes(ethPub.Type(), ethPub.Bytes()[:5])
		require.Error(t, err)
		require.ErrorAs(t, err, &invLen)
	} else {
		// should error if type known but not enabled
		_, err := keys.PubKeyFromCometTypeAndBytes(secp256k1eth.KeyType, []byte{})
		require.Error(t, err)
	}

	// --- bls keys are not enabled in comet and should fail ---
	_, err = bls12381.GenPrivKey()
	require.Error(t, err)

	// --- unsupported type ---
	_, err = keys.PubKeyFromCometTypeAndBytes("not-a-key", []byte{1, 2, 3})
	require.Error(t, err)
	var unsup encoding.ErrUnsupportedKey
	require.ErrorAs(t, err, &unsup)
}
