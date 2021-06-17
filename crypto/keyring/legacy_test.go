package keyring_test

import (
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestNewLegacyKeyBase(t *testing.T) {
	dir := t.TempDir()
	encCfg := simapp.MakeTestEncodingConfig()

	kb, err := keyring.NewLegacy("keybasename", dir, encCfg.Marshaler)
	require.NoError(t, err)
	require.NoError(t, kb.Close())
}

func TestLegacyKeybase(t *testing.T) {
	dir := t.TempDir()
	encCfg := simapp.MakeTestEncodingConfig()

	// Backup testdata
	require.NoError(t, copy.Copy("testdata", dir))

	kb, err := keyring.NewLegacy("keys", filepath.Join(dir, "keys"), encCfg.Marshaler)
	require.NoError(t, err)
	t.Cleanup(func() { kb.Close() })

	keys, err := kb.List()
	require.NoError(t, err)
	require.Equal(t, 2, len(keys))

	armor, err := kb.ExportPubKey(keys[0].Name)
	require.NoError(t, err)
	require.NotEmpty(t, armor)

	_, err = kb.ExportPrivKey(keys[0].Name, "12345678", "12345678")
	require.Error(t, err)

	armoredInfo, err := kb.Export(keys[0].Name)
	require.NoError(t, err)
	require.NotEmpty(t, armoredInfo)
}
