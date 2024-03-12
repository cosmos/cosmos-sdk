package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/x/epochs/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	Keeper struct {
		cdc         codec.BinaryCodec
		environment appmodule.Environment
		hooks       types.EpochHooks

		Schema    collections.Schema
		EpochInfo collections.Map[string, types.EpochInfo]
	}
)

// NewKeeper returns a new keeper by codec and storeKey inputs.
func NewKeeper(env appmodule.Environment, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
	k := Keeper{
		cdc:         cdc,
		environment: env,
		EpochInfo:   collections.NewMap(sb, types.KeyPrefixEpoch, "epoch_info", collections.StringKey, codec.CollValue[types.EpochInfo](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Set the gamm hooks.
func (k Keeper) SetHooks(eh types.EpochHooks) Keeper {
	if k.hooks != nil {
		panic("cannot set epochs hooks twice")
	}

	k.hooks = eh

	return k
}

func (k Keeper) Logger() log.Logger {
	return k.environment.Logger
}
