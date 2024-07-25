package appdatasim

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
)

func TestDiffAppData_1(t *testing.T) {
	appSim, err := NewSimulator(Options{
		AppSchema: schematesting.ExampleAppSchema,
	})
	require.NoError(t, err)

	mirror, err := NewSimulator(Options{
		// just one module
		AppSchema: map[string]schema.ModuleSchema{
			"all_kinds": schematesting.ExampleAppSchema["all_kinds"],
		},
	})
	require.NoError(t, err)

	blockData := appSim.BlockDataGen().Example(1)
	require.NoError(t, appSim.ProcessBlockData(blockData))

	golden.Assert(t, DiffAppData(appSim, mirror), "diff_example.txt")
}
