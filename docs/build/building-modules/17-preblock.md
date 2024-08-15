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

* It runs before the `BeginBlocker` of all modules
* It can modify consensus parameters in storage, and signal the caller through the return value.

Modules are required to get the consensus params from the consensus module. Consensus params located in `sdk.Context` were deprecated and should be treated as unsafe. `sdk.Context` is deprecated due to it being a global state within the entire state machine, it has been replaced with `appmodule.Environment`.
