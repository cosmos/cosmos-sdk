# Schema & Indexer Testing

This module contains core test utilities and fixtures for testing `cosmossdk.io/schema` and `cosmossdk.io/schema/indexer` functionality. It is managed as a separate go module to manage versions better and allow for dependencies on useful testing libraries without imposing those elsewhere.

The two primary intended uses of this library are:
- testing that indexers can handle all valid app data that they might be asked to index
- testing that state management frameworks properly map their data to and from schema types

## Testing Indexers

Indexers are expected to process all valid `schema` and `appdata` types, yet it may be hard to find a data set in the wild that comprehensively represents the full valid range of these types. This library provides utilities for simulating such data. The simplest way for an indexer to leverage this test framework is to implement the `appdatasim.HasAppData` type against their data store. Then the `appdatasim.Simulator` can be used to generate deterministically random valid data that can be sent to the indexer and also stored in the simulator. After each generated block is applied, `appdatasim.DiffAppData` can be used to compare the expected state in the simulator to the actual state in the indexer.

This example code shows how this might look in a test:

```go
func TestMyIndexer(t *testing.T) {
	var myIndexerListener appdata.Listener
	var myIndexerAppData appdatasim.HasAppData
    // do the setup for your indexer and return an appdata.Listener to consume updates and the appdatasim.HasAppData instance to check the actual vs expected data
    myIndexerListener, myIndexerAppData := myIndexer.Setup() 
	
    simulator, err := appdatasim.NewSimulator(appdatatest.SimulatorOptions{
        AppSchema: indexertesting.ExampleAppSchema,
        StateSimOptions: statesim.Options{
            CanRetainDeletions: true,
        },
		Listener: myIndexerListener,
    })
    require.NoError(t, err)
    
    blockDataGen := simulator.BlockDataGen()
    for i := 0; i < 1000; i++ {
		// using Example generates a deterministic data set based
        // on a seed so that regression tests can be created OR rapid.Check can
        // be used for fully random property-based testing
        blockData := blockDataGen.Example(i)
		
        // process the generated block data with the simulator which will also
        // send it to the indexer
        require.NoError(t, simulator.ProcessBlockData(blockData))
		
        // compare the expected state in the simulator to the actual state in the indexer and expect the diff to be empty
		require.Empty(t, appdatasim.DiffAppData(simulator, myIndexerAppData))
    }
}
```

## Testing State Management Frameworks

The compliance of frameworks like `cosmossdk.io/collections` and `cosmossdk.io/orm` with `cosmossdk.io/schema` can be tested with this framework. One example of how this might be done is if there is a `KeyCodec` that represents an array of `schema.Field`s then `schematesting.ObjectKeyGen` might be used to generate a random object key which encoded and then decoded and then `schematesting.DiffObjectKeys` is used to compare the expected key with the decoded key. If such state management frameworks require that users that schema compliance when implementing things like `KeyCodec`s then those state management frameworks should specify best practices for users.