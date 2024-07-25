package appdatasim

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
)

// this test checks that diffs in app data are deterministic and can be used for regression testing
func TestDiffAppData(t *testing.T) {
	appSim, err := NewSimulator(Options{
		AppSchema: schematesting.ExampleAppSchema,
	})
	require.NoError(t, err)

	mirror, err := NewSimulator(Options{
		// add just one module to the mirror
		AppSchema: map[string]schema.ModuleSchema{
			"all_kinds": schematesting.ExampleAppSchema["all_kinds"],
		},
	})
	require.NoError(t, err)

	// mirror one block
	blockGen := appSim.BlockDataGenN(50, 100)
	blockData := blockGen.Example(1)
	require.NoError(t, appSim.ProcessBlockData(blockData))
	require.NoError(t, mirror.ProcessBlockData(blockData))

	// produce another block, but don't mirror it so that they're out of sync
	blockData = blockGen.Example(2)
	require.NoError(t, appSim.ProcessBlockData(blockData))

	golden.Assert(t, DiffAppData(appSim, mirror), "diff_example.txt")
}
