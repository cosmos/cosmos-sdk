package appdatasim

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/schema/testing"
)

func TestAppSimulator_ExampleSchema(t *testing.T) {
	out := &bytes.Buffer{}
	appSim := NewSimulator(SimulatorOptions{
		AppSchema: schematesting.ExampleAppSchema,
		Listener:  WriterListener(out),
	})

	require.NoError(t, appSim.Initialize())

	blockDataGen := appSim.BlockDataGen()

	for i := 0; i < 10; i++ {
		data := blockDataGen.Example(i + 1)
		require.NoError(t, appSim.ProcessBlockData(data))
	}

	golden.Assert(t, out.String(), "app_sim_example_schema.txt")
}
