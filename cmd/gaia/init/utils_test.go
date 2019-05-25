package init

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/stretchr/testify/require"
)

func TestExportGenesisFileWithTime(t *testing.T) {
	t.Parallel()
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()

	fname := filepath.Join(dir, "genesis.json")
	require.NoError(t, ExportGenesisFileWithTime(fname, "test", nil, json.RawMessage(""), time.Now()))
}

func TestLoadGenesisDoc(t *testing.T) {
	t.Parallel()
	dir, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()

	fname := filepath.Join(dir, "genesis.json")
	require.NoError(t, ExportGenesisFileWithTime(fname, "test", nil, json.RawMessage(""), time.Now()))

	_, err := LoadGenesisDoc(codec.Cdc, fname)
	require.NoError(t, err)

	// Non-existing file
	_, err = LoadGenesisDoc(codec.Cdc, "non-existing-file")
	require.Error(t, err)

	malformedFilename := filepath.Join(dir, "malformed")
	malformedFile, err := os.Create(malformedFilename)
	require.NoError(t, err)
	fmt.Fprint(malformedFile, "invalidjson")
	malformedFile.Close()
	// Non-existing file
	_, err = LoadGenesisDoc(codec.Cdc, malformedFilename)
	require.Error(t, err)
}
