package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	tcrypto "github.com/tendermint/tendermint/crypto"
)

func TestRealLedgerSecp256k1(t *testing.T) {

	if os.Getenv("WITH_LEDGER") == "" {
		t.Skip("Set WITH_LEDGER to run code on real ledger")
	}
	msg := []byte("{\"account_number\":\"3\",\"chain_id\":\"1234\",\"fee\":{\"amount\":[{\"amount\":\"150\",\"denom\":\"atom\"}],\"gas\":\"5000\"},\"memo\":\"memo\",\"msgs\":[[\"%s\"]],\"sequence\":\"6\"}")

	path := DerivationPath{44, 60, 0, 0, 0}

	priv, err := NewPrivKeyLedgerSecp256k1(path)
	require.Nil(t, err, "%+v", err)
	pub := priv.PubKey()
	sig, err := priv.Sign(msg)
	require.Nil(t, err)

	valid := pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// now, let's serialize the public key and make sure it still works
	bs := priv.PubKey().Bytes()
	pub2, err := tcrypto.PubKeyFromBytes(bs)
	require.Nil(t, err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig, err = priv.Sign(msg)
	require.Nil(t, err)
	valid = pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = pub.Bytes()
	bpub, err := tcrypto.PubKeyFromBytes(bs)
	require.NoError(t, err)
	require.Equal(t, pub, bpub)
}

// TestRealLedgerErrorHandling calls. These tests assume
// the ledger is not plugged in....
func TestRealLedgerErrorHandling(t *testing.T) {
	if os.Getenv("WITH_LEDGER") != "" {
		t.Skip("Skipping on WITH_LEDGER as it tests unplugged cases")
	}

	// first, try to generate a key, must return an error
	// (no panic)
	path := DerivationPath{44, 60, 0, 0, 0}
	_, err := NewPrivKeyLedgerSecp256k1(path)
	require.Error(t, err)
}
