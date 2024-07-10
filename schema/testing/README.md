# Schema Testing

This module contains core test utilities and fixtures for testing `comossdk.io/schema` and `cosmossdk.io/schema/indexer` functionality. It is managed as a separate go module to manage versions better and allow for dependencies on useful testing libraries without imposing those elsewhere.

## Indexer Testing

The primary intended use of this library is for testing indexer implementations. It may be hard to find a data set in the wild that comprehensively represents the full range of supported `schema` and `appdata` types. This library provides utilities for simulating the full range of all types of data that indexers should support. The example code below demonstrates a recommended way for indexers to leverage this simulated data:

```go
func TestMyIndexer(t *testing.T) {
    indexerListener := myIndexer.Setup()
    simulator := appdatasim.NewSimulator(appdatatest.SimulatorOptions{
        AppSchema: indexertesting.ExampleAppSchema,
        StateSimOptions: statesim.Options{
            CanRetainDeletions: true,
        },
		Listener: indexerListener,
    })
    
    require.NoError(t, simulator.Initialize())
    
    blockDataGen := fixture.BlockDataGen()
    for i := 0; i < 1000; i++ {
		// using Example generates a deterministic data set based
        // on the seed, different seeds can be used and rapid.Check can
        // be used for full property based testing
        blockData := blockDataGen.Example(i)
        require.NoError(t, fixture.ProcessBlockData(blockData))
    
        require.NoError(t, fixture.AppState().ScanModules(func    (moduleName string, modState *statesim.Module) error {
			// check that the expected state in each statesim.Module
            // matches the actual state in the indexed database
        })
    }
}
```