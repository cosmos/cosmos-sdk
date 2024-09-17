package appdatasim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/schema/appdata"
	schematesting "cosmossdk.io/schema/testing"
	"cosmossdk.io/schema/testing/statesim"
)

func TestAppSimulator_mirror(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		testAppSimulatorMirror(t, false)
	})
	t.Run("retain deletes", func(t *testing.T) {
		testAppSimulatorMirror(t, true)
	})
}

func testAppSimulatorMirror(t *testing.T, retainDeletes bool) { //nolint: thelper // this isn't a test helper function
	stateSimOpts := statesim.Options{CanRetainDeletions: retainDeletes}
	mirror, err := NewSimulator(Options{
		StateSimOptions: stateSimOpts,
	})
	require.NoError(t, err)

	appSim, err := NewSimulator(Options{
		AppSchema: schematesting.ExampleAppSchema,
		Listener: appdata.PacketForwarder(func(packet appdata.Packet) error {
			return mirror.ProcessPacket(packet)
		}),
		StateSimOptions: stateSimOpts,
	})
	require.NoError(t, err)

	blockDataGen := appSim.BlockDataGenN(50, 100)

	for i := 0; i < 10; i++ {
		data := blockDataGen.Example(i + 1)
		require.NoError(t, appSim.ProcessBlockData(data))
		require.Empty(t, DiffAppData(appSim, mirror))
	}
}

func TestAppSimulator_exampleSchema(t *testing.T) {
	out := &bytes.Buffer{}
	appSim, err := NewSimulator(Options{
		AppSchema:       schematesting.ExampleAppSchema,
		Listener:        writerListener(out),
		StateSimOptions: statesim.Options{},
	})
	require.NoError(t, err)

	blockDataGen := appSim.BlockDataGenN(10, 20)

	for i := 0; i < 10; i++ {
		data := blockDataGen.Example(i + 1)
		require.NoError(t, appSim.ProcessBlockData(data))
	}

	// we do a golden test so that we can have some human-readable proof that
	// the simulator is emitting updates that look like what we expect
	// make sure you check the golden tests when the simulator is changed
	// this can be updated by running "go test . -update"
	golden.Assert(t, out.String(), "app_sim_example_schema.txt")
}

// writerListener returns a listener that writes to the provided writer. It currently
// only covers callbacks which are called by the simulator, but others will be added
// as the simulator covers other cases.
func writerListener(w io.Writer) appdata.Listener {
	return appdata.Listener{
		StartBlock: func(data appdata.StartBlockData) error {
			_, err := fmt.Fprintf(w, "StartBlock: %v\n", data)
			return err
		},
		OnTx:     nil,
		OnEvent:  nil,
		OnKVPair: nil,
		Commit: func(data appdata.CommitData) (completionCallback func() error, err error) {
			_, err = fmt.Fprintf(w, "Commit: %v\n", data)
			return nil, err
		},
		InitializeModuleData: func(data appdata.ModuleInitializationData) error {
			bz, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "InitializeModuleData: %s\n", bz)
			return err
		},
		OnObjectUpdate: func(data appdata.ObjectUpdateData) error {
			bz, err := json.Marshal(data)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "OnObjectUpdate: %s\n", bz)
			return err
		},
	}
}
