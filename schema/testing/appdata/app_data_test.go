package appdatatest

import (
	"bytes"
	"testing"

	"gotest.tools/v3/golden"

	"cosmossdk.io/schema/testing"
)

func TestAppSimulator_ExampleSchema(t *testing.T) {
	out := &bytes.Buffer{}
	appSim := NewAppSimulator(t, AppSimulatorOptions{
		AppSchema:          schematesting.ExampleAppSchema,
		Listener:           WriterListener(out),
		MaxUpdatesPerBlock: 20,
	})

	appSim.Initialize()

	for i := 0; i < 10; i++ {
		appSim.NextBlock()
	}

	golden.Assert(t, out.String(), "app_sim_example_schema.txt")
}
