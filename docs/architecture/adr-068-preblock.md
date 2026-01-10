# ADR 068: Preblock

## Changelog

* Sept 13, 2023: Initial Draft

## Status

DRAFT

## Abstract

Introduce `PreBlock`, which runs before the begin blocker of other modules, and allows modifying consensus parameters, and the changes are visible to the following state machine logics.

## Context

When upgrading to SDK 0.47, the storage format for consensus parameters changed, but in the migration block, `ctx.ConsensusParams()` is always `nil` because it fails to load the old format using new code. It's supposed to be migrated by the `x/upgrade` module first, but unfortunately, the migration happens in `BeginBlocker` handler, which runs after the `ctx` is initialized.
When we try to solve this, we find the `x/upgrade` module can't modify the context to make the consensus parameters visible for the other modules. The context is passed by value, and the SDK team wants to keep it that way, which is good for isolation between modules.

## Alternatives

The first alternative solution introduced a `MigrateModuleManager`, which only includes the `x/upgrade` module right now, and baseapp will run their `BeginBlocker`s before the other modules, and reload context's consensus parameters in between.

## Decision

Suggested this new lifecycle method.

### `PreBlocker`

There are two semantics around the new lifecycle method:

* It runs before the `BeginBlocker` of all modules
* It can modify consensus parameters in storage, and signal the caller through the return value.

When it returns `ConsensusParamsChanged=true`, the caller must refresh the consensus parameters in the finalize context:

```
app.finalizeBlockState.ctx = app.finalizeBlockState.ctx.WithConsensusParams(app.GetConsensusParams())
```

The new ctx must be passed to all the other lifecycle methods.


## Consequences

### Backwards Compatibility

### Positive

### Negative

### Neutral

## Further Discussions

## Test Cases

## References

* [1] https://github.com/cosmos/cosmos-sdk/issues/16494
* [2] https://github.com/cosmos/cosmos-sdk/pull/16583
* [3] https://github.com/cosmos/cosmos-sdk/pull/17421
* [4] https://github.com/cosmos/cosmos-sdk/pull/17713
