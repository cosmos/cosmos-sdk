package keyring_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
)

func TestNewLegacyKeyBase(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)
	kb, err := keyring.NewLegacy("keybasename", dir)
	require.NoError(t, err)
	require.NoError(t, kb.Close())
}
