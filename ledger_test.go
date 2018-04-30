package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealLedger(t *testing.T) {

	if os.Getenv("WITH_LEDGER") == "" {
		t.Skip("Set WITH_LEDGER to run code on real ledger")
	}
	msg := []byte("kuhehfeohg")

	priv, err := NewPrivKeyLedgerSecp256k1()
	require.Nil(t, err, "%+v", err)
	pub := priv.PubKey()
	sig := priv.Sign(msg)

	valid := pub.VerifyBytes(msg, sig)
	assert.True(t, valid)

	// now, let's serialize the key and make sure it still works
	bs := priv.Bytes()
	priv2, err := PrivKeyFromBytes(bs)
	require.Nil(t, err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	pub2 := priv2.PubKey()
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig = priv2.Sign(msg)
	valid = pub.VerifyBytes(msg, sig)
	assert.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = pub.Bytes()
	bpub, err := PubKeyFromBytes(bs)
	require.NoError(t, err)
	assert.Equal(t, pub, bpub)
}

// TestRealLedgerErrorHandling calls. These tests assume
// the ledger is not plugged in....
func TestRealLedgerErrorHandling(t *testing.T) {
	if os.Getenv("WITH_LEDGER") != "" {
		t.Skip("Skipping on WITH_LEDGER as it tests unplugged cases")
	}

	// first, try to generate a key, must return an error
	// (no panic)
	_, err := NewPrivKeyLedgerSecp256k1()
	require.Error(t, err)
}
