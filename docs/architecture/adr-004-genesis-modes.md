# ADR 4: Multiple Genesis modes

## Changelog

- 07/26/2019: First draft

## Context

### Problem

Currently, each some of the modules `GenesisState` are populated if the corresponding
object passed through the Genesis JSON is empty (_i.e_ as an empty array) or if it's
undefined or the value is zero.

This leads to confussion as you migh pass a valid genesis file which contains zero/empty
values and have `InitGenenesis` perform this uncessesary check.

### Proposed Solution

The proposed solution is to have two ways of initializing the daemon:

1. **New (empty) chain**: the user provides the minimum amount of genesis fields required to
start a chain. The rest is loaded directly from the mandatory genesis parameters.
2. **Restarted chain**: the user provides all the fields from genesis.

If an empty chain from `1.` recives _any_ field that should be populated, the initialization
of the app should `panic`. Same applies for `2.` if a single field is missing.

The fields that are not required are the ones in general are checked with simulation
invariants.

## Decision

Add a flag `FlagPopulateGenesis = "populate-genesis"` on `server/start.go` that is
passed to the `AppCreator` (eg: `NewSimApp` or `NewGaiaApp`) when `<app>d start` is called.
This tells the app to fill the missing genesis properties with the mandatory fields. For eg:

```go
app := NewSimApp(logger, db, nil, true, 0, SetPopulateGenesis(populate))
```

Additionally, a new option function needs to be passed as parameter to the `BaseApp` during its
initiallization:

```go
// ./baseapp/options.go
//
// SetPopulateGenesis sets an option to populate or not the missing values
// of the each module GenesisState from the mandatory fields given on the genesis file
func SetPopulateGenesis(pupulate bool) func(*BaseApp) {
  return func(bap *BaseApp) { bap.setPopulateGenesis(populate)}
}
```

Which requires the following additions to `BaseApp`:

```go
// ./baseapp/baseapp.go
type BaseApp struct {
  // ...
  populateGenesis bool // flag for pupulating the genesis state from the mandatory fields on genesis.json
  // ...
}
```

```go
// ./baseapp/baseapp.go
func (app *BaseApp) setPopulateGenesis(populate bool) {
  app.populateGenesis = populate
}
```

The new flag is then used on the app `InitChainer` function, which calls:

```go
// add populateGenesis parameter
func (am AppModule) InitGenesis(ctx sdk.Context, populateGenesis bool, data json.RawMessage)
```

Then, every module has to be updated to follow each mode. For example, on `supply/genesis.go`,
[`InitGenesis`](https://github.com/cosmos/cosmos-sdk/blob/0ba74bb4b77f465e4c661552381732d8612e7c0b/x/supply/genesis.go#L12)
needs to be modified to the following:

```go
func InitGenesis(ctx sdk.Context, keeper Keeper, ak types.AccountKeeper, populate bool, data GenesisState) {
  if populate {
    if !data.Supply.Total.Empty() {
      panic("total supply should be initialized with empty coins ('[]') on populate genesis mode")
    }

    ak.IterateAccounts(ctx,
      func(acc authtypes.Account) (stop bool) {
        data.Supply.Total = data.Supply.Total.Add(acc.GetCoins())
        return false
      },
    )
  }

  keeper.SetSupply(ctx, data.Supply)
}
```

## Status

Proposed

## Consequences

### Positive

- Reduce confusion when initialize genesis as now there're two clearly distinct
modes for it

### Negative

### Neutral

- Add an additional parameter to `BaseApp`
- Apps must update their genesis files and tests to follow one mode or the other

## References

- [#2862](https://github.com/cosmos/cosmos-sdk/issues/2862)
- [#4568](https://github.com/cosmos/cosmos-sdk/issues/4568)
