# ADR 4: Multiple Genesis modes

## Changelog

- July 26, 2019: Initial draft

## Context

### Problem

Currently, some of the modules `GenesisState` are populated only if the corresponding
object passed through the Genesis JSON is empty (_i.e_ as an empty array) or if it's
undefined or the value is zero.

This leads to confusion as you might pass a valid genesis file which contains zero/empty
values and have `InitGenenesis` perform this unnecessary check.

### Proposed Solution

The proposed solution is to add a command to update a genesis file with the
missing values/field.

With this command, the user provides a `genesis.json` file with the minimum amount
of genesis fields required to start a chain. The rest is loaded directly from the
mandatory genesis parameters. In general, the genesis fields that are populated
are the ones checked by the simulation invariants.

If the given genesis file contains _any_ field that shouldn't be populated, the
execution should `panic` and exit. Similarly, the initialization of the app (_i.e_
`<app>d init`) should `panic` if any genesis field is undefined.

The flow of this process can be summarized in 3 steps:

  1. Import genesis file with missing fields
  2. Initialize Genesis with flag to update missing values
  3. Export updated genesis file

## Decision

Create an seperate `populate-genesis` command on `x/genutils` that performs the 3 steps described above.

To archieve this, every module needs to be able to generate the missing fields from the existing state. For example, on `x/supply/genesis.go`,
[`InitGenesis`](https://github.com/cosmos/cosmos-sdk/blob/0ba74bb4b77f465e4c661552381732d8612e7c0b/x/supply/genesis.go#L12)
needs to be modified to the following:

```go
// add populateGenesis parameter
func InitGenesis(ctx sdk.Context, keeper Keeper, ak types.AccountKeeper, data GenesisState, populateGenesis bool) {
  if populateGenesis {
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

As this is part of the `AppModule` pattern, we also need to update its `InitGenesis`:

```go
// ./x/<module_name>/module.go
//
// add populateGenesis parameter
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage, populateGenesis bool) {
  InitGenesis(ctx, am.k, am.ak ,data, populateGenesis) // supply module InitGenesis
}
```

Upstream, the application module `Manager` calls each module's `InitGenesis` function:

```go
// ./types/module/module.go
//
// InitGenesis performs init genesis functionality for modules. If the populateGenesis boolean is true, it updates the missing values of each module's GenesisState from the mandatory fields given in the genesis file data.
func (m *Manager) InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage, populateGenesis bool) abci.ResponseInitChain {
  var validatorUpdates []abci.ValidatorUpdate
  for _, moduleName := range m.OrderInitGenesis {
    if genesisData[moduleName] == nil {
      continue
    }
    moduleValUpdates := m.Modules[moduleName].InitGenesis(ctx, genesisData[moduleName], populate)

    // validator updates
  }
  return abci.ResponseInitChain{
    Validators: validatorUpdates,
  }
}
```

Finally, the `pupulate-genesis` command imports the provided genesis dile and passes the parameter to the module `Manager` call to update the state and then export it:

```go
// TODO: params
func PopulateCmd(ctx *Context, appCreator AppCreator) *cobra.Command {
    cmd := &cobra.Command{
    Use:   "pupulate-genesis [path/to/genesis.json]",
    Short: "Update the missing non-mandatory values from the genesis file",
    Long:  `Update the missing non-mandatory values from the genesis file by
    `,
    Args:  cobra.ExactArgs(1),
    RunE: func(_ *cobra.Command, args []string) error {
      // initialize args
      genFile = args[0]

      genFile := config.GenesisFile()

      // import genesis
      appState, err := codec.MarshalJSONIndent(cdc, mbm.DefaultGenesis())
      if err != nil {
        return errors.Wrap(err, "Failed to marshall default genesis state")
      }

      genDoc := &types.GenesisDoc{}
      if _, err := os.Stat(genFile); err != nil {
        if !os.IsNotExist(err) {
          return err
        }
      } else {
        genDoc, err = types.GenesisDocFromFile(genFile)
        if err != nil {
          return errors.Wrap(err, "Failed to read genesis doc from file")
        }
      }

      // create app
      // TODO: create app
      app := appCreator()

      err = app.Codec().UnmarshalJSON(appState, &genesisState)
      if err != nil {
        panic(err)
      }

      // update app
      app.mm.InitGenesis(ctx, genesisState, true)


      // export updates state
      // TODO: export
      if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
        return errors.Wrap(err, "Failed to export genesis file")
      },
    }

    return cmd
}
```

## Status

Proposed

## Consequences

### Positive

- Reduce confusion when initializing the app, as now you can only provide a genesis file with all its fields defined.
- Add a command to populate a genesis file from the mandatory values.

### Negative

- The chain now only starts if all the fields of the genesis file are populated.
- Remove the current single option to partially fill in the genesis on the initialization (_i.e_ only with a few non-mandatory fields missing).
This can be replaced by the following flow:

  1. Use the new empty chain option to auto-generate the genesis fields
  2. Modify the fields desired before starting the chain

### Neutral

- Add an additional parameter to `BaseApp`
- Apps must update their genesis files and tests to follow one mode or the other

## References

- [#2862](https://github.com/cosmos/cosmos-sdk/issues/2862)
- [#4568](https://github.com/cosmos/cosmos-sdk/issues/4568)
