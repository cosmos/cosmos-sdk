---
sidebar_position: 1
---

# PreBlocker

:::note Synopsis
`PreBlocker` is optional method module developers can implement in their module. They will be triggered before [`BeginBlock`](../../learn/advanced/00-baseapp.md#beginblock).
:::

:::note Pre-requisite Readings

* [Module Manager](./01-module-manager.md)

:::

## PreBlocker

There are two semantics around the new lifecycle method:

- It runs before the `BeginBlocker` of all modules
- It can modify consensus parameters in storage, and signal the caller through the return value.

When it returns `ConsensusParamsChanged=true`, the caller must refresh the consensus parameter in the deliver context:
```
app.finalizeBlockState.ctx = app.finalizeBlockState.ctx.WithConsensusParams(app.GetConsensusParams())
```

The new ctx must be passed to all the other lifecycle methods.

<!-- TODO: leaving this here to update docs with core api changes  -->
