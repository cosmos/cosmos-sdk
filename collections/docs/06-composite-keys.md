# Composite keys

So far we've worked only with simple keys, like uint64, the account address, etc.
There are some more complex cases in which we need to deal with composite keys.

Composite keys are keys composed of multiple keys. How we can efficiently iterate
over composite keys, etc.

# Example

In our example we will show-case how we can use collections when we are dealing with balances, similar to bank,
a balance is a mapping between `(address, denom) => math.Int` the composite key in our case is `(address, denom)`.

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


## Working with composite key collections

Let's expand on the example we used before:

````go
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

func (k Keeper) SetBalance(ctx sdk.Context, address sdk.AccAddress, denom string, amount math.Int) error {
	key := collections.Join(address, denom)
	return k.Balances.Set(ctx, key, amount)
}

func (k Keeper) GetBalance(ctx sdk.Context, address sdk.AccAddress, denom string) (math.Int, error) {
	return k.Balances.Get(ctx, collections.Join(address, denom))
}

func (k Keeper) GetAllAddressBalances(ctx sdk.Context, address sdk.AccAddress) (sdk.Coins, error) {
	balances := sdk.NewCoins()

	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](address)

	iter, err := k.Balances.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}

	kvs, err := iter.KeyValues()
	if err != nil {
		return nil, err
	}

	for _, kv := range kvs {
		balances = balances.Add(sdk.NewCoin(kv.Key.K2(), kv.Value))
	}
	return balances, nil
}

func (k Keeper) GetAllAddressBalancesBetween(ctx sdk.Context, address sdk.AccAddress, startDenom, endDenom string) (sdk.Coins, error) {
    rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](address).
        StartInclusive(startDenom).
        EndInclusive(endDenom)

    iter, err := k.Balances.Iterate(ctx, rng)
    if err != nil {
        return nil, err
	}
    ...
}
````

### SetBalance

As we can see here we're setting the balance of an address for a specific denom.
We use the `collections.Join` function to generate the composite key.
`collections.Join` returns a `collections.Pair` (which is the key of our `collections.Map`)

`collections.Pair` contains the two keys we have joined, it also exposes  two methods: `K1` to fetch the 1st part of the
key and `K2` to fetch the second part.

The what we do, as always, we use the `collections.Map.Set` method to map the composite key to our value (`math.Int`in this case)

### GetBalance

It shows the same thing how we can get a value in composite key collection, we simply use `collections.Join` to compose the key.

### GetAllAddressBalances

We show here how we can use `collections.PrefixedPairRange` to iterate over all the keys starting with the provided address.
Concretely the iteration will report all the balances belonging to the provided address.

The first part is that we instantiate a `PrefixedPairRange`, which is a `Ranger` implementer aimed to help
in `Pair` keys iterations.

```go
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](address)
```

As we can see here we're passing the type parameters of the `collections.Pair` because golang type inference
with respect to generics is not as permissive as other languages, so we need to explitly say what are the types of the pair key.

### GetAllAddressesBalancesBetween

This showcases how we can further specialise our range to limit the results further, by specifying 
the range between the second part of the key (in our case the denoms, which are strings).

