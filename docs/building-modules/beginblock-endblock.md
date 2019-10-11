# BeginBlocker and EndBlocker

## Pre-requisite Reading

- [Module Manager](./module-manager.md)

## Synopsis

`BeginBlocker` and `EndBlocker` are optional methods module developers can implement in their module. They will be triggered at the beginning and at the end of each block respectively. 

## BeginBlocker and EndBlocker

`BeginBlocker` and `EndBlocker` are a way for module developers to add automatic execution of logic to their module. This is a powerful tool that should be used carefully, as complex automatic functions can slow down or even halt the chain. 

When needed, `BeginBlocker` and `EndBlocker` are implemented as part of the [`AppModule` interface](./module-manager.md#appmodule). The `BeginBlock` and `EndBlock` methods of the interface implemented in `module.go` generally defer to `BeginBlocker` and `EndBlocker` methods respectively, which are usually implemented in a `abci.go` file. 

The actual implementation of `BeginBlocker` and `EndBlocker` are very similar to that of a [`handler`](./handler.md):

- They generally use the [`keeper`](./keeper.md) and [`ctx`](../core/context.md) to retrieve information about the latest state. 
- If needed, they use the `keeper` and `ctx` to trigger state-transitions. 
- If needed, they can emit [`events`](../core/events.md) via the `ctx`'s `EventManager`. 

A specificity of the `EndBlocker` is that it can return validator updates to the underlying consensus engine in the form of an [`[]abci.ValidatorUpdates`](https://tendermint.com/docs/app-dev/abci-spec.html#validatorupdate). 

For more, see an [example implementation of `BeginBlocker` from the `distr` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/distribution/abci.go) and an [example implementation of `EndBlocker`]( https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/handler.go#L44)

## Next

Learn about [`keeper`s](./keeper.md).
