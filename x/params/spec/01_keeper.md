# Keeper

In the app initialization stage, `Keeper.Subspace(Paramspace)` is passed to the user modules, and the subspaces are stored in `Keeper.spaces`. Later it can be retrieved with `Keeper.GetSubspace`, so the keepers holding `Keeper` can access to any subspace. For example, Gov module can take `Keeper` as its argument and modify parameter of any subspace when a `ParameterChangeProposal` is accepted.  

Example:

```go
type MasterKeeper struct {
	pk params.Keeper
}

func (k MasterKeeper) SetParam(ctx sdk.Context, space string, key string, param interface{}) {
	space, ok := k.ps.GetSubspace(space)
	if !ok {
		return
	}
	space.Set(ctx, key, param)
}
```
