`blockstm` implements the [block-stm algorithm](https://arxiv.org/abs/2203.06871), it follows the paper pseudocode pretty closely.

The main API is a simple function call `ExecuteBlock`:

```golang
type ExecuteFn func(TxnIndex, MultiStore)
func ExecuteBlock(
 ctx context.Context,           // context for cancellation
 blockSize int,                 // the number of the transactions to be executed
 stores []storetypes.StoreKey,  // the list of store keys to support
 storage MultiStore,            // the parent storage, after all transactions are executed, the whole change sets are written into parent storage at once
 executors int,                 // how many concurrent executors to spawn
 executeFn ExecuteFn,           // callback function to actually execute a transaction with a wrapped `MultiStore`.
) error
```

Broken internal invariants will cause panics.

The main deviations from the paper are:

### Optimisation

We applied the optimization described in section 4 of the paper:

```
Block-STM calls add_dependency from the VM itself, and can thus re-read and continue execution when false is returned.
```

When the VM execution reads an `ESTIMATE` mark, it'll hang on a `CondVar`, so it can resume execution after the dependency is resolved,
much more efficient than abortion and rerun.

### Support Deletion, Iteration, and MultiStore

These features are necessary for integration with cosmos-sdk.

The multi-version data structure is implemented with nested btree for easier iteration support,
the `WriteSet` is also implemented with a btree, and it takes advantage of ordered property to optimize some logic.

The internal data structures are also adapted with multiple stores in mind.

### Attribution

This package was originally authored in [go-block-stm](https://github.com/crypto-org-chain/go-block-stm). We have brought the full source tree into the SDK so that we can natively incorporate the library and required changes into the SDK. Over time we expect to incorporate optimizations and deviations from the upstream implementation.
