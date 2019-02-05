package crypto

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/stretchr/testify/require"
)

// This tests assume a ledger is not plugged in
func TestLedgerErrorHandling(t *testing.T) {
	// first, try to generate a key, must return an error
	// (no panic)
	path := *hd.NewParams(44, 555, 0, false, 0)
	_, err := NewPrivKeyLedgerSecp256k1(path)
	require.Error(t, err)
}
