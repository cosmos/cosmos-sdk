//go:build system_test

package systemtests

import (
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
	"path/filepath"
	"testing"
)

func TestChainExportImport(t *testing.T) {
	// Scenario:
	//   given: a state dump from a running chain
	//   when: new chain is initialized with exported state
	//   then: the chain should start and produce blocks
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	sut.StartChain(t)
	sut.AwaitNBlocks(t, 2)
	sut.StopChain()

	outFile := filepath.Join(t.TempDir(), "exported_genesis.json")
	cli.RunCommandWithArgs("genesis", "export", "--home="+sut.NodeDir(0), "--output-document="+outFile)
	exportedContent, err := os.ReadFile(outFile)
	require.NoError(t, err)
	exportedState := gjson.Get(string(exportedContent), "app_state").Raw

	sut.ModifyGenesisJSON(t, func(genesis []byte) []byte {
		state, err := sjson.SetRawBytes(genesis, "app_state", []byte(exportedState))
		require.NoError(t, err)
		return state
	})
	sut.StartChain(t)
	sut.AwaitNBlocks(t, 2)
}
