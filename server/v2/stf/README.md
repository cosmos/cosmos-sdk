# State Transition Function (STF)

STF is a function that takes a state and an action as input and returns the next state. It does not assume the execution model of the application nor consensus.

The state transition function receives a read only instance of state. It does not directly write to disk, instead it will return the state changes which has undergone within the application. The state transition function is deterministic, meaning that given the same input, it will always produce the same output.

## BranchDB

BranchDB is a cache of all the reads done within a block, simulation or transaction validation. It takes a read-only instance of state and creates its own write instance using a btree. After all state transitions are done, the new change sets are returned to the caller.

The BranchDB can be replaced and optimized for specific use cases. The implementation is as follows

```go
   type branchdb func(state store.ReaderMap) store.WriterMap
```

## GasMeter

GasMeter is a utility that keeps track of the gas consumed by the state transition function. It is used to limit the amount of computation that can be done within a block.

The GasMeter can be replaced and optimized for specific use cases. The implementation is as follows:

```go
type (
 // gasMeter is a function type that takes a gas limit as input and returns a gas.Meter.
 // It is used to measure and limit the amount of gas consumed during the execution of a function.
 gasMeter func(gasLimit uint64) gas.Meter

 // wrapGasMeter is a function type that wraps a gas meter and a store writer map.
 wrapGasMeter func(meter gas.Meter, store store.WriterMap) store.WriterMap
)
```

THe wrapGasMeter is used in order to consume gas. Application developers can seamlsessly replace the gas meter with their own implementation in order to customize consumption of gas.
