package keyring_test

import (
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
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

func TestLegacyKeybase(t *testing.T) {
	dir, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	// Backup testdata
	require.NoError(t, copy.Copy("testdata", dir))

	kb, err := keyring.NewLegacy("keys", filepath.Join(dir, "keys"))
	require.NoError(t, err)
	t.Cleanup(func() { kb.Close() })

	keys, err := kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keys))

	armor, err := kb.ExportPubKey(keys[0].GetName())
	require.NoError(t, err)
	require.NotEmpty(t, armor)

	armoredInfo, err := kb.Export(keys[0].GetName())
	require.NoError(t, err)
	require.NotEmpty(t, armoredInfo)
}
