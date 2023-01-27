# IndexedMap

`collections.IndexedMap` is a collection that uses under the hood a `collections.Map`, and has a struct
which contains the indexes that we need to define.

# Example

Let's say we have an `auth.BaseAccount` struct which looks like the following:

```go
type BaseAccount struct {
	AccountNumber uint64     `protobuf:"varint,3,opt,name=account_number,json=accountNumber,proto3" json:"account_number,omitempty"`
	Sequence      uint64     `protobuf:"varint,4,opt,name=sequence,proto3" json:"sequence,omitempty"`
}
```

First of all when we save our accounts in state we map them using a primary key, in our it's the `sdk.AccAddress`.
If it were to be a `collections.Map` it would be `collections.Map[sdk.AccAddres, authtypes.BaseAccount]`.

Then we also want to be able to get an account not only by its `sdk.AccAddress`, but also by its `AccountNumber`.

So we can say we want to create an `Index` that maps our `BaseAccount` to its `AccountNumber`.

We also know that this `Index` is unique. Unique means that there can only be one `BaseAccount` that maps to a specific 
`AccountNumber`.

First of all we start by defining the object that contains our index:

```go
var AccountsNumberIndexPrefix = collections.NewPrefix(1)

type AccountsIndexes struct {
	Number *indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
}

func (a AccountsIndexes) IndexesList() []collections.Index[sdk.AccAddress, authtypes.BaseAccount] {
	return []collections.Index[sdk.AccAddress, authtypes.BaseAccount]{a.Number}
}

func NewAccountIndexes(sb *collections.SchemaBuilder) AccountsIndexes {
	return AccountsIndexes{
		Number: indexes.NewUnique(
			sb, AccountsNumberIndexPrefix, "accounts_by_number",
			collections.Uint64Key, sdk.AccAddressKey,
			func(_ sdk.AccAddress, v authtypes.BaseAccount) (uint64, error) {
				return v.AccountNumber, nil
			},
		),
	}
}
```

We create an `AccountIndexes` struct which contains a field: `Number`. This field represents our `AccountNumber` index.
`AccountNumber` is a field of `authtypes.BaseAccount` and it's a `uint64`.

Then we can see in our `AccountIndexes` struct the `Number` field is defined as:

```go
*indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
```

Where the first type parameter is `uint64`, which is the field type of our index.
The second type parameter is the primary key `sdk.AccAddress`
And the third type parameter is the actual object we're storing `authtypes.BaseAccount`.

Then we implement a function called `IndexesList` on our `AccountIndexes` struct, this will be used
by the `IndexedMap` to keep the underlying map in sync with the indexes, in our case `Number`.
This function just needs to return the slice of indexes contained in the struct.

Then we create a `NewAccountIndexes` function that instantiates and returns the `AccountsIndexes` struct.

The function takes a `SchemaBuilder`. Then we instantiate our `indexes.Unique`, let's analyse the arguments we pass to
`indexes.NewUnique`.

#### Instantiating a `indexes.Unique`

The first three arguments, we already know them, they are: `SchemaBuilder`, `Prefix` which is our index prefix (the partition 
where index keys relationship for the `Number` index will be maintained), and the human name for the `Number` index.

The second argument is a `collections.Uint64Key` which is a key codec to deal with `uint64` keys, we pass that because
the key we're trying to index is a `uint64` key (the account number), and then we pass as fifth argument the primary key codec, 
which in our case is `sdk.AccAddress` (remember: we're mapping `sdk.AccAddress` => `BaseAccount`).

Then as last parameter we pass a function that: given the `BaseAccount` returns its `AccountNumber`.

After this we can proceed instantiating our `IndexedMap`.

```go
var AccountsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema   collections.Schema
	Accounts *collections.IndexedMap[sdk.AccAddress, authtypes.BaseAccount, AccountsIndexes]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewIndexedMap(
			sb, AccountsPrefix, "accounts",
			sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc),
			NewAccountIndexes(sb),
		),
	}
}
```

As we can see here what we do, for now, is the same thing as we did for `collections.Map`.
We pass it the `SchemaBuilder`, the `Prefix` where we plan to store the mapping between `sdk.AccAddress` and `authtypes.BaseAccount`,
the human name and the respective `sdk.AccAddress` key codec and `authtypes.BaseAccount` value codec.

Then we pass the instantiation of our `AccountIndexes` through `NewAccountIndexes`.

Full example:

```go
package docs

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var AccountsNumberIndexPrefix = collections.NewPrefix(1)

type AccountsIndexes struct {
	Number *indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
}

func (a AccountsIndexes) IndexesList() []collections.Index[sdk.AccAddress, authtypes.BaseAccount] {
	return []collections.Index[sdk.AccAddress, authtypes.BaseAccount]{a.Number}
}

func NewAccountIndexes(sb *collections.SchemaBuilder) AccountsIndexes {
	return AccountsIndexes{
		Number: indexes.NewUnique(
			sb, AccountsNumberIndexPrefix, "accounts_by_number",
			collections.Uint64Key, sdk.AccAddressKey,
			func(_ sdk.AccAddress, v authtypes.BaseAccount) (uint64, error) {
				return v.AccountNumber, nil
			},
		),
	}
}

var AccountsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema   collections.Schema
	Accounts *collections.IndexedMap[sdk.AccAddress, authtypes.BaseAccount, AccountsIndexes]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewIndexedMap(
			sb, AccountsPrefix, "accounts",
			sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc),
			NewAccountIndexes(sb),
		),
	}
}
```

## Working with IndexedMaps

Whilst instantiating `collections.IndexedMap` is tedious, working with them is extremely smooth.

Let's take the full example, and expand it with some use-cases.

```go
package docs

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var AccountsNumberIndexPrefix = collections.NewPrefix(1)

type AccountsIndexes struct {
	Number *indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
}

func (a AccountsIndexes) IndexesList() []collections.Index[sdk.AccAddress, authtypes.BaseAccount] {
	return []collections.Index[sdk.AccAddress, authtypes.BaseAccount]{a.Number}
}

func NewAccountIndexes(sb *collections.SchemaBuilder) AccountsIndexes {
	return AccountsIndexes{
		Number: indexes.NewUnique(
			sb, AccountsNumberIndexPrefix, "accounts_by_number",
			collections.Uint64Key, sdk.AccAddressKey,
			func(_ sdk.AccAddress, v authtypes.BaseAccount) (uint64, error) {
				return v.AccountNumber, nil
			},
		),
	}
}

var AccountsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema   collections.Schema
	Accounts *collections.IndexedMap[sdk.AccAddress, authtypes.BaseAccount, AccountsIndexes]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewIndexedMap(
			sb, AccountsPrefix, "accounts",
			sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc),
			NewAccountIndexes(sb),
		),
	}
}

func (k Keeper) CreateAccount(ctx sdk.Context, addr sdk.AccAddress) error {
	nextAccountNumber := k.getNextAccountNumber()
	
	newAcc := authtypes.BaseAccount{
		AccountNumber: nextAccountNumber,
		Sequence:      0,
	}
	
	return k.Accounts.Set(ctx, addr, newAcc)
}

func (k Keeper) RemoveAccount(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.Accounts.Remove(ctx, addr)
} 

func (k Keeper) GetAccountByNumber(ctx sdk.Context, accNumber uint64) (sdk.AccAddress, authtypes.BaseAccount, error) {
	accAddress, err := k.Accounts.Indexes.Number.MatchExact(ctx, accNumber)
	if err != nil {
		return nil, authtypes.BaseAccount{}, err
	}
	
	acc, err := k.Accounts.Get(ctx, accAddress)
	return accAddress, acc, nil
}

func (k Keeper) GetAccountsByNumber(ctx sdk.Context, startAccNum, endAccNum uint64) ([]authtypes.BaseAccount, error) {
	rng := new(collections.Range[uint64]).
		StartInclusive(startAccNum).
		EndInclusive(endAccNum)
	
	iter, err := k.Accounts.Indexes.Number.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	
	return indexes.CollectValues(ctx, k.Accounts, iter)
}


func (k Keeper) getNextAccountNumber() uint64 {
	return 0
}
```


