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
	Schema collections.Schema
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
