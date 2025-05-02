## Custom `MintFn` with external dependencies

```go
type Deps struct {
	Gov govkeeper.Keeper
}

func CustomMintFn(d Deps) types.MintFn {
	return func(ctx sdk.Context, k keeper.Keeper) error {
		vp := d.Gov.GetVotingPower(ctx)         // external keeper
		k.Mint(ctx, vp.QuoInt64(100))           // example logic
		return nil
	}
}
```

В `app/app.go`:

```go
mintMod := mintmodule.NewAppModule(
	appCodec, mintKeeper, ak, bk,
	mintmodule.WithMintFn(CustomMintFn(deps)),
)
``` 