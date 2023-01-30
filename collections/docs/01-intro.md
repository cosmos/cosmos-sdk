## Core types

Collections offers 5 different APIs to work with state, which will be explored in the next sections, these APIs are: 
- ``Map``: to work with typed arbitrary KV pairings.
- ``KeySet``: to work with just typed keys
- ``Item``: to work with just one typed value
- ``Sequence``: which is a monotonically increasing number.
- ``IndexedMap``: which combines ``Map`` and `KeySet` to provide a `Map` with indexing capabilities.

## Preliminary components

Before exploring the different collections types and their capability it is necessary to introduce
the three components that every collection shares. In fact when instantiating a collection type by doing, for example,
```collections.NewMap/collections.NewItem/...``` you will find yourself having to pass them some common arguments.

For example, in code:

```go
package collections

import (
    "cosmossdk.io/collections"
    storetypes "cosmossdk.io/store/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

var AllowListPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema    collections.Schema
	AllowList collections.KeySet[string]
}

func NewKeeper(storeKey *storetypes.KVStoreKey) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))

	return Keeper{
		AllowList: collections.NewKeySet(sb, AllowListPrefix, "allow_list", collections.StringKey),
	}
}

```

Let's analyse the shared arguments, what they do, and why we need them.

### SchemaBuilder

The first argument passed is the ``SchemaBuilder``

Schema builder is a structure that keeps track of all the state of a module, it is not required by the collections
themselve to deal with state, but it offers a dynamic and reflective way for clients to explore a module's state.

We instantiate a ``SchemaBuilder`` by passing it a function that given the modules store key returns the module's specific store.

We then need to pass the schema builder to every collection type we instantiate in our keeper, in our case the `AllowList`.

### Prefix

The second argument passed to our ``KeySet`` is a `collections.Prefix`, a prefix represents a partition of the module's KVStore
where all the state of a specific collection will be saved. Since a module can have multiple collections, example:

- Module params will become a `collections.Item`, 
- The `AllowList` is a `collections.KeySet`.
- etc.

We don't want a collection to write over the state of the other so we pass it a prefix, which defines this storage partition
owned by the collection.

If you already built modules, the prefix translates to the items you were creating in your ``types/keys.go`` file, example: https://github.com/cosmos/cosmos-sdk/blob/main/x/feegrant/key.go#L27

your old:
```go
var (
	// FeeAllowanceKeyPrefix is the set of the kvstore for fee allowance data
	// - 0x00<allowance_key_bytes>: allowance
	FeeAllowanceKeyPrefix = []byte{0x00}

	// FeeAllowanceQueueKeyPrefix is the set of the kvstore for fee allowance keys data
	// - 0x01<allowance_prefix_queue_key_bytes>: <empty value>
	FeeAllowanceQueueKeyPrefix = []byte{0x01}
)
```
becomes:
```go
var (
	// FeeAllowanceKeyPrefix is the set of the kvstore for fee allowance data
	// - 0x00<allowance_key_bytes>: allowance
	FeeAllowanceKeyPrefix = collections.NewPrefix(0)

	// FeeAllowanceQueueKeyPrefix is the set of the kvstore for fee allowance keys data
	// - 0x01<allowance_prefix_queue_key_bytes>: <empty value>
	FeeAllowanceQueueKeyPrefix = collections.NewPrefix(1)
)
```

#### Rules

``collections.NewPrefix`` accepts either `uint8`, `string` or `[]bytes` it's good practice to use an always increasing `uint8`for disk space efficiency.

A collection **MUST NOT** share the same prefix as another collection in the same module, and a collection prefix **MUST NEVER** start with the same prefix as another, examples:

```go
prefix1 := collections.NewPrefix("prefix")
prefix2 := collections.NewPrefix("prefix") // THIS IS BAD!
```

```go
prefix1 := collections.NewPrefix("a")
prefix2 := collections.NewPrefix("aa") // prefix2 starts with the same as prefix1: BAD!!!
```
### Human-Readable Name

The third parameter we pass to a collection is a string, which is a human-readable name.
It is needed to make the role of a collection understandable by clients who have no clue about
what a module is storing in state.

#### Rules

Each collection in a module **MUST** have a unique humanised name.

## Key and Value Codecs

A collection is generic over the type you can use as keys or values.
Which makes collections dumb, but also means that hypothetically we can store everything
that can be a go type into a collection. We are not bounded to any type of encoding (be it proto, json or whatever)

So a collection needs to be given a way to understand how to convert your keys and values to bytes. 
This is achieved through ``KeyCodec`` and `ValueCodec`, which are arguments that you pass to your 
collections when you're instantiating them using the ```collections.NewMap/collections.NewItem/...``` 
instantiation functions.

NOTE: Generally speaking you will never be required to implement your own ``Key/ValueCodec`` as 
the SDK and collections libraries already come with default, safe and fast implementation of those.
You might need to implement them only if you're migrating to collections and there are state layout incompatibilities.

Let's explore an example:

````go
package collections

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var IDsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema    collections.Schema
	IDs   collections.Map[string, uint64]
}

func NewKeeper(storeKey *storetypes.KVStoreKey) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))

	return Keeper{
		IDs: collections.NewMap(sb, IDsPrefix, "ids", collections.StringKey, collections.Uint64Value),
	}
}
````

We're now instantiating a map where the key is string and the value is uint64.
We already know the first three arguments of the ``NewMap`` function.

The fourth parameter is our `KeyCodec`, we know that the ``Map`` has `string` as key so we pass it a `KeyCodec` that handles strings as keys.

The fifth parameter is our `ValueCodec`, we know that the `Map` as a `uint64` as value so we pass it a `ValueCodec` that handles uint64.

Collections already comes with all the required implementations for golang primitive types.

Let's make another example, this falls closer to what we build using cosmos SDK, let's say we want
to create a `collections.Map` that maps account addresses to their base account. So we want to map an `sdk.AccAddress` to an `auth.BaseAccount` (which is a proto):

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
	Schema    collections.Schema
	Accounts   collections.Map[sdk.AccAddress, authtypes.BaseAccount]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewMap(sb, AccountsPrefix, "accounts",
			sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc)),
	}
}
```

As we can see here since our `collections.Map` maps `sdk.AccAddress` to `authtypes.BaseAccount`,
we use the `sdk.AccAddressKey` which is the `KeyCodec` implementation for `AccAddress` and we use `codec.CollValue` to
encode our proto type `BaseAccount`.

Generally speaking you will always find the respective key and value codecs for types in the go.mod path you're using
to import that type. If you want to encode proto values refer to the codec `codec.CollValue` function which allows you
to encode any type implement the proto.Message interface.


