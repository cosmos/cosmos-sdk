# Map

We analyse the first and most important collection type, the ``collections.Map``.
This is the type that everything else builds on top of.

## Use case

A `collections.Map` is used to map arbitrary keys with arbitrary values. 

## Example

It's easier to explain a `collections.Map` capabilities through an example:

```go
package collections

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"fmt"
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

func (k Keeper) CreateAccount(ctx sdk.Context, addr sdk.AccAddress, account authtypes.BaseAccount) error {
	has, err := k.Accounts.Has(ctx, addr)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("account already exists: %s", addr)
	}
	
	err = k.Accounts.Set(ctx, addr, account)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) (authtypes.BaseAccount, error) {
	acc, err := k.Accounts.Get(ctx, addr)
	if err != nil {
		return authtypes.BaseAccount{}, err
	}
	
	return acc,	nil
}

func (k Keeper) RemoveAccount(ctx sdk.Context, addr sdk.AccAddress) error {
	err := k.Accounts.Remove(ctx, addr)
	if err != nil {
		return err
	}
	return nil
}
```

### Set method

Set maps the provided `AccAddress` (the key) to the `auth.BaseAccount` (the value). 

Under the hood the `collections.Map` will conver the key and value to bytes using the [key and value codec](README.md#key-and-value-codecs).
It will prepend to our bytes key the [prefix](README.md#prefix) and store it in the KVStore of the module.

### Has method

The has method reports if the provided key exists in the store.

### Get method

The get method accepts the `AccAddress` and returns the associated `auth.BaseAccount` if it exists, otherwise it errors.

### Remove method

The remove method accepts the `AccAddress` and removes it from the store. It won't report errors
if it does not exist, to check for existence before removal use the ``Has`` method.

### Iteration

Iteration has a separate section.