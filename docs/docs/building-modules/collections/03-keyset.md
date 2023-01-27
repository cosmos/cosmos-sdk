# KeySet

The second type of collection is `collections.KeySet`, as the word suggests it maintains
only a set of keys without values.

#### Implementation curiosity

A `collections.KeySet` is just a `collections.Map` with a `key` but no value.
The value internally is always the same and is represented as an empty byte slice ```[]byte{}```.

## Example

As always we explore the collection type through an example:

```go
package collections

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ValidatorsSetPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema        collections.Schema
	ValidatorsSet collections.KeySet[sdk.ValAddress]
}

func NewKeeper(storeKey *storetypes.KVStoreKey) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		ValidatorsSet: collections.NewKeySet(sb, ValidatorsSetPrefix, "validators_set", sdk.ValAddressKey),
	}
}

func (k Keeper) AddValidator(ctx sdk.Context, validator sdk.ValAddress) error {
	has, err := k.ValidatorsSet.Has(ctx, validator)
	if err != nil {
		return err
	}
	if has {
		return fmt.Errorf("validator already in set: %s", validator)
	}
	
	err = k.ValidatorsSet.Set(ctx, validator)
	if err != nil {
		return err
	}
	
	return nil
}

func (k Keeper) RemoveValidator(ctx sdk.Context, validator sdk.ValAddress) error {
	err := k.ValidatorsSet.Remove(ctx, validator)
	if err != nil {
		return err
	}
	return nil
}
```
The first difference we notice is that `KeySet` needs use to specify only one type parameter: the key (`sdk.ValAddress` in this case).
The second difference we notice is that `KeySet` in its `NewKeySet` function does not require
us to specify a `ValueCodec` but only a `KeyCodec`. This is because a `KeySet` only saves keys and not values.

Let's explore the methods.

### Has method

Has allows us to understand if a key is present in the `collections.KeySet` or not, functions in the same way as `collections.Map.Has
`

### Set method

Set inserts the provided key in the `KeySet`.

### Remove method

Remove removes the provided key from the `KeySet`, it does not error if the key does not exist, 
if existence check before removal is required it needs to be coupled with the `Has` method.
