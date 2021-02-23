<!--
order: 1
-->

# Keeper

In the app initialization stage, [subspaces](02_subspace.md) can be allocated for other modules' keeper using `Keeper.Subspace` and are stored in `Keeper.spaces`. Then, those modules can have a reference to their specific parameter store through `Keeper.GetSubspace`.

Example:

```go
type ExampleKeeper struct {
	paramSpace paramtypes.Subspace
}

func (k ExampleKeeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
```
