package crypto

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

var ledgerEnabledEnv = "TEST_WITH_LEDGER"

func TestRealLedgerSecp256k1(t *testing.T) {
	if os.Getenv(ledgerEnabledEnv) == "" {
		t.Skip(fmt.Sprintf("Set '%s' to run code on a real ledger", ledgerEnabledEnv))
	}

	msg := []byte("kuhehfeohg")
	path := DerivationPath{44, 60, 0, 0, 0}

	priv, err := NewPrivKeyLedgerSecp256k1(path)
	require.Nil(t, err, "%s", err)

	pub := priv.PubKey()
	sig, err := priv.Sign(msg)
	require.Nil(t, err)

	valid := pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// now, let's serialize the key and make sure it still works
	bs := priv.Bytes()
	priv2, err := tmcrypto.PrivKeyFromBytes(bs)
	require.Nil(t, err, "%s", err)

	// make sure we get the same pubkey when we load from disk
	pub2 := priv2.PubKey()
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig, err = priv2.Sign(msg)
	require.Nil(t, err)
	valid = pub.VerifyBytes(msg, sig)
	require.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = pub.Bytes()
	bpub, err := tmcrypto.PubKeyFromBytes(bs)
	require.NoError(t, err)
	require.Equal(t, pub, bpub)
}

// TestRealLedgerErrorHandling calls. These tests assume
// the ledger is not plugged in....
func TestRealLedgerErrorHandling(t *testing.T) {
	if os.Getenv(ledgerEnabledEnv) == "" {
		t.Skip(fmt.Sprintf("Set '%s' to run code on a real ledger", ledgerEnabledEnv))
	}

	// first, try to generate a key, must return an error
	// (no panic)
	path := DerivationPath{44, 60, 0, 0, 0}
	_, err := NewPrivKeyLedgerSecp256k1(path)
	require.Error(t, err)
}
