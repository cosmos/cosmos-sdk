# Item

The third type of collection is the `collections.Item`.
It stores only one single item, it's useful for example for parameters, there's only one instance
of parameters in state always.

#### implementation curiosity

A `collections.Item` is just a `collections.Map` with no key but just a value.
The key is the prefix of the collection!

## Example

```go
package collections

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var ParamsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema        collections.Schema
	Params collections.Item[stakingtypes.Params]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Params: collections.NewItem(sb, ParamsPrefix, "params", codec.CollValue[stakingtypes.Params](cdc)),
	}
}

func (k Keeper) UpdateParams(ctx sdk.Context, params stakingtypes.Params) error {
	err := k.Params.Set(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetParams(ctx sdk.Context) (stakingtypes.Params, error) {
	return k.Params.Get(ctx)
}
```

The first key difference we notice is that we specify only one type parameter, which is the value we're storing.
The second key difference is that we don't specify the `KeyCodec`, since we store only one item we already know the key
and the fact that it is constant.


