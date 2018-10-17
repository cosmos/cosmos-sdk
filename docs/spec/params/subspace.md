# Subspace

## Basic Usage

First, declare parameter space and parameter keys for the module. Then include params.Subspace in the keeper. Since we prefix the keys with the spacename, it is recommended to use the same name with the module's.

```go
const (
	DefaultParamspace = "mymodule"
)

const (
	KeyParameter1 = "myparameter1"
	KeyParameter2 = "myparameter2"
)

type Keeper struct {
	cdc *wire.Codec
	key sdk.StoreKey

	ps params.Subspace
}
```

Pass a params.Subspace to NewKeeper with DefaultParamSubspace (or another)

```go
app.myKeeper = mymodule.NewKeeper(cdc, key, app.paramStore.SubStore(mymodule.DefaultParamspace))
```

`NewKeeper` should register a `TypeTable`, which defines a map from parameter keys from types.

```go
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, space params.Subspace) Keeper {
    return Keeper {
        cdc: cdc,
        key: key,
        ps: space.WithTypeTable(ParamTypeTable()),
    }
}
```

Now we can access to the paramstore using Paramstore Keys

```go
var param MyStruct
k.ps.Get(KeyParameter1, &param)
k.ps.Set(KeyParameter2, param)
```

# Genesis Usage

Declare a struct for parameters and make it implement params.ParamSet. It will then be able to be passed to SetParamSet.

```go
type MyParams struct {
	Parameter1 uint64
	Parameter2 string
}

// Implements params.ParamSet
// KeyValuePairs must return the list of (ParamKey, PointerToTheField)
func (p *MyParams) KeyValuePairs() params.KeyValuePairs {
	return params.KeyFieldPairs {
		{KeyParameter1, &p.Parameter1},
		{KeyParameter2, &p.Parameter2},
	}
}

func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.ps.SetParamSet(ctx, &data.params)
}
```

The method is pointer receiver because there could be a case that we read from the store and set the result to the struct.

