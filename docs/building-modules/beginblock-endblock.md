<!--
order: 6
synopsis: "`BeginBlocker` and `EndBlocker` are optional methods module developers can implement in their module. They will be triggered at the beginning and at the end of each block respectively, when the [`BeginBlock`](../core/baseapp.md#beginblock) and [`EndBlock`](../core/baseapp.md#endblock) ABCI messages are received from the underlying consensus engine."
-->

# BeginBlocker and EndBlocker

## Pre-requisite Readings {hide}

- [Module Manager](./module-manager.md) {prereq}

## BeginBlocker and EndBlocker

`BeginBlocker` and `EndBlocker` are a way for module developers to add automatic execution of logic to their module. This is a powerful tool that should be used carefully, as complex automatic functions can slow down or even halt the chain. 

When needed, `BeginBlocker` and `EndBlocker` are implemented as part of the [`AppModule` interface](./module-manager.md#appmodule). The `BeginBlock` and `EndBlock` methods of the interface implemented in `module.go` generally defer to `BeginBlocker` and `EndBlocker` methods respectively, which are usually implemented in a **`abci.go`** file. 

The actual implementation of `BeginBlocker` and `EndBlocker` in `./abci.go` are very similar to that of a [`handler`](./handler.md):

- They generally use the [`keeper`](./keeper.md) and [`ctx`](../core/context.md) to retrieve information about the latest state. 
- If needed, they use the `keeper` and `ctx` to trigger state-transitions. 
- If needed, they can emit [`events`](../core/events.md) via the `ctx`'s `EventManager`. 

A specificity of the `EndBlocker` is that it can return validator updates to the underlying consensus engine in the form of an [`[]abci.ValidatorUpdates`](https://tendermint.com/docs/app-dev/abci-spec.html#validatorupdate). This is the preferred way to implement custom validator changes. 

It is possible for developers to defined the order of execution between the `BeginBlocker`/`EndBlocker` functions of each of their application's modules via the module's manager `SetOrderBeginBlocker`/`SetOrderEndBlocker` methods. For more on the module manager, click [here](./module-manager.md#manager). 

See an example implementation of `BeginBlocker` from the `distr` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/distribution/abci.go#L10-L32

and an example implementation of `EndBlocker` from the `staking` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/staking/handler.go#L44-L96

## Next {hide}

Learn about [`keeper`s](./keeper.md) {hide}
