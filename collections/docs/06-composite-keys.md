# Composite keys

So far we've worked only with simple keys, like uint64, the account address, etc.
There are some more complex cases in which we need to deal with composite keys.

Composite keys are keys composed of multiple keys. How we can efficiently iterate
over composite keys, etc.

# Example

In our example we will show-case how we can use collections when we are dealing with balances, similar to bank,
a balance is a mapping between `(address, denom) => Int` the composite key in our case is `(address, denom)`.

## Instantiation of a composite key collection

```go
package collections

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


var BalancesPrefix = collections.NewPrefix(1)

type Keeper struct {
	Schema   collections.Schema
	Balances collections.Map[collections.Pair[sdk.AccAddress, string], math.Int]
}

func NewKeeper(storeKey *storetypes.KVStoreKey) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Balances: collections.NewMap(
			sb, BalancesPrefix, "balances",
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
			math.IntValue,
		),
	}
}
```

### The Map Key definition

First of all we can see that in order to define a composite key of two elements we use the `collections.Pair` type:
````go
collections.Map[collections.Pair[sdk.AccAddress, string], math.Int]
````

`collections.Pair` defines a key composed of two other keys, in our case the first part is `sdk.AccAddress`, the second
part is `string`.

### The Key Codec instantiation

The arguments to instantiate are always the same, the only thing that changes is how we instantiate
the ``KeyCodec``, since this key is composed of two keys we use `collections.PairKeyCodec` which generates
a `KeyCodec` composed of two key codecs. The first one will encode the first part of the key, the second one will
encode the second part of the key.




