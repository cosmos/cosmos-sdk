package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

type Keeper struct {
	storeService store.KVStoreService

	cdc   codec.BinaryCodec
	hooks types.EpochHooks

	Schema    collections.Schema
	EpochInfo collections.Map[string, types.EpochInfo]
}

// NewKeeper returns a new keeper by codec and storeKey inputs.
func NewKeeper(storeService store.KVStoreService, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		EpochInfo:    collections.NewMap(sb, types.KeyPrefixEpoch, "epoch_info", collections.StringKey, codec.CollValue[types.EpochInfo](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// SetHooks sets the hooks on the x/epochs keeper.
func (k *Keeper) SetHooks(eh types.EpochHooks) {
	if k.hooks != nil {
		panic("cannot set epochs hooks twice")
	}

	k.hooks = eh
}
