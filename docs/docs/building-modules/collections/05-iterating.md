# Iteration

One of the key features of the ``KVStore`` is iterating over keys.

Collections which deal with keys (so `Map`, `KeySet` and `IndexedMap`) allow you to iterate
over keys in a safe and typed way. They all share the same API, the only difference being
that ``KeySet`` returns a different type of `Iterator` because `KeySet` only deals with keys.

NOTE: that every collection shares the same `Iterator` semantics.

Let's have a look at the `Map.Iterate` method:

```go
func (m Map[K, V]) Iterate(ctx context.Context, ranger Ranger[K]) (Iterator[K, V], error) 
```

It accepts a `collections.Ranger[K]` which an API that instructs map on how to iterate over keys.
As always we don't need to implement anything here as `collections` already provides some generic `Ranger` implementers
that expose all you need to work with ranges.

## Example

We have a collections.Map that maps accounts using uint64 IDs.

```go
package collections

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var AccountsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema   collections.Schema
	Accounts collections.Map[uint64, authtypes.BaseAccount]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewMap(sb, AccountsPrefix, "params", collections.Uint64Key, codec.CollValue[authtypes.BaseAccount](cdc)),
	}
}

func (k Keeper) GetAllAccounts(ctx sdk.Context) ([]authtypes.BaseAccount, error) {
	// passing a nil Ranger equals to: iterate over every possible key
	iter, err := k.Accounts.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	accounts, err := iter.Values()
	if err != nil {
		return nil, err
	}

	return accounts, err
}

func (k Keeper) IterateAccountsBetween(ctx sdk.Context, start, end uint64) ([]authtypes.BaseAccount, error) {
	// The collections.Range API offers a lot of capabilities
	// like defining where the iteration starts or ends.
	rng := new(collections.Range[uint64]).
		StartInclusive(start).
		EndExclusive(end).
		Descending()

	iter, err := k.Accounts.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	accounts, err := iter.Values()
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (k Keeper) IterateAccounts(ctx sdk.Context, do func(id uint64, acc authtypes.BaseAccount) (stop bool)) error {
	iter, err := k.Accounts.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if do(kv.Key, kv.Value) {
			break
		}
	}
	return nil
}
```

Let's analyse each method in the example and how it makes use of the `Iterate` and the returned `Iterator` API.

### GetAllAccounts

In `GetAllAccounts` we pass to our `Iterate` a nil `Ranger`. This means that the returned `Iterator` will include
all the existing keys within the collection.

Then we use some the `Values` method from the returned `Iterator` API to collect all the values into a slice.

`Iterator` offers other methods such as `Keys()` to collect only the keys and not the values and `KeyValues` to collect
all the keys and values.


### IterateAccountsBetween

Here we make use of the `collections.Range` helper to specialise our range.
We make it start in a point through `StartInclusive` and end in the other with `EndExclusive`, then
we instruct it to report us results in reverse order through `Descending`

Then we pass the range instruction to `Iterate` and get an `Iterator` which will contain only the results
we specified in the range.

Then we use again th `Values` method of the `Iterator` to collect all the results.

`collections.Range` also offers a `Prefix` API which is not 

### IterateAccounts

Here we showcase how to lazily collect values from an Iterator. Note that `Keys/Values/KeyValues` fully consume
and close the `Iterator`, here we need to explictly do a `defer iterator.Close()` call.

`Iterator` also exposes a `Value` and `Key` method to collect only the current value or key, if collecting both is not needed.

