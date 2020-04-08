package keyring_test

import (
	"io"
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

	_, err = kb.ExportPrivKey(keys[0].GetName(), "12345678", "12345678")
	require.Error(t, err)

	armoredInfo, err := kb.Export(keys[0].GetName())
	require.NoError(t, err)
	require.NotEmpty(t, armoredInfo)

	importer, err := keyring.NewInfoImporter("cosmos", "memory", "", nil)
	require.NoError(t, err)
	err = importer.Import("test", "")
	require.Error(t, err)
	require.Equal(t, io.EOF, err)
	require.NoError(t, importer.Import("test", armoredInfo))

	err = importer.Import("test", armoredInfo)
	require.Error(t, err)
	require.Equal(t, `public key already exist in keybase`, err.Error())
}
