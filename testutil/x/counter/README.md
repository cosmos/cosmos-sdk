---
sidebar_position: 1
---

# `x/counter`

Counter is a module used for testing purposes within the Cosmos SDK.

## Hooks

The module supports hook injection using `depinject`. Follow these steps to implement hooks:

1. Define the hook interface and a wrapper implementing `depinject.OnePerModuleType`:
```go
type CounterHooks interface {
	AfterIncreaseCount(ctx context.Context, newCount int64) error
}

type CounterHooksWrapper struct{ CounterHooks }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (CounterHooksWrapper) IsOnePerModuleType() {}

```
2. Add a `CounterHooks` field to the keeper:
```go
type Keeper struct{
	//...
	hooks CounterHooks
}

func (k *Keeper) SetHooks(hooks ...CounterHooks){
	k.hooks = hooks
}
```
3. Create a `depinject` invoker function
```go
func InvokeSetHooks(keeper *keeper.Keeper, counterHooks map[string]types.CounterHooksWrapper) error {
	//...
    keeper.SetHooks(multiHooks)
    return nil
}

```
4. Inject the hooks during app initialization:
```go
appConfig = appconfig.Compose(&appv1alpha1.Config{
    Modules: []*appv1alpha1.ModuleConfig{
        // ....
        {
            Name:   types.ModuleName,
            Config: appconfig.WrapAny(&types.Module{}),
        },
    }
})
appConfig = depinject.Configs(
    AppConfig(),
    runtime.DefaultServiceBindings(),
    depinject.Supply(
        logger,
        viper,
        map[string]types.CounterHooksWrapper{
            "counter": types.CounterHooksWrapper{&types.Hooks{}},
        },
))

```
By following these steps, you can successfully integrate hooks into the `x/counter` module using `depinject`.
