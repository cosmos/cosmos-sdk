---
sidebar_position: 1
---

# PreBlocker

:::note Synopsis
`PreBlocker` is an optional method module developers can implement in their module. They will be triggered before [`BeginBlock`](../../learn/advanced/00-baseapp.md#beginblock).
:::

:::note Pre-requisite Readings

* [Module Manager](./01-module-manager.md)

:::

## PreBlocker

There are two semantics around the new lifecycle method:

* It runs before the `BeginBlocker` of all modules
* It can modify consensus parameters in storage, and signal the caller through the return value.

When it returns `ConsensusParamsChanged=true`, the caller must refresh the consensus parameters in the finalize context and update the block gas meter accordingly. Conceptually, this looks like:

```
ctx = ctx.WithConsensusParams(app.GetConsensusParams(ctx))
gasMeter := app.getBlockGasMeter(ctx)
ctx = ctx.WithBlockGasMeter(gasMeter)
```

The new finalize context must then be used for all the other lifecycle methods in the block.

<!-- TODO: leaving this here to update docs with core api changes  -->
