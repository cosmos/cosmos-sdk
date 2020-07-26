package genutil

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/KiraCore/cosmos-sdk/testutil"

	"github.com/stretchr/testify/require"
)

func TestExportGenesisFileWithTime(t *testing.T) {
	t.Parallel()
	dir, cleanup := testutil.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	fname := filepath.Join(dir, "genesis.json")
	require.NoError(t, ExportGenesisFileWithTime(fname, "test", nil, json.RawMessage(""), time.Now()))
}
