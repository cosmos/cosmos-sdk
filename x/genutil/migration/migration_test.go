package migration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var oldGenFilePath = "./testdata/old_app_genesis.json"

func TestMigration(t *testing.T) {
	tempDir := t.TempDir()

	// clean all content on this directory
	err := os.RemoveAll(tempDir)
	require.NoError(t, err)

	migrator, err := NewMigrator(oldGenFilePath, tempDir)
	require.NoError(t, err)

	// should not be able to get app genesis from new genesis file
	// since validators address are still in hex string and not cons address
	_, err = types.AppGenesisFromFile(oldGenFilePath)
	require.ErrorContains(t, err, "error unmarshalling AppGenesis: decoding bech32 failed")

	err = migrator.MigrateGenesisFile()
	require.NoError(t, err)

	// should be able to get app genesis from new genesis file
	_, err = types.AppGenesisFromFile(tempDir)
	require.NoError(t, err)
}
