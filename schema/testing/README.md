# Schema & Indexer Testing

This module contains core test utilities and fixtures for testing `comossdk.io/schema` and `cosmossdk.io/schema/indexer` functionality. It is managed as a separate go module to manage versions better and allow for dependencies on useful testing libraries without imposing those elsewhere.

The two primary intended uses of this library are:
- testing that indexers can handle all valid app data that they might be asked to index
- testing that state management frameworks (such as collections or orm) properly map their data to and from schema types

## Testing Indexers

Indexers are expected to process all valid `schema` and `appdata` types, yet it may be hard to find a data set in the wild that comprehensively represents the full valid range of these types. This library provides utilities for simulating such data. The example code below demonstrates one way indexers may leverage this simulated data to test compliance:

```go
func TestMyIndexer(t *testing.T) {
    indexerListener := myIndexer.Setup()
    simulator, err := appdatasim.NewSimulator(appdatatest.SimulatorOptions{
        AppSchema: indexertesting.ExampleAppSchema,
        StateSimOptions: statesim.Options{
            CanRetainDeletions: true,
        },
		Listener: indexerListener,
    })
    
    require.NoError(t, err)
    
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

## Testing State Management Frameworks

More information on this will be added as these capabilities are developed.