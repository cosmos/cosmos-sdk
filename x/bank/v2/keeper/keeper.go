package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/x/bank/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper defines the bank/v2 module keeper.
// All fields are not exported, as they should only be accessed through the module's.
type Keeper struct {
	authority    []byte
	addressCodec address.Codec
	environment  appmodulev2.Environment
	schema       collections.Schema
	params       collections.Item[types.Params]
}

func NewKeeper(authority []byte, addressCodec address.Codec, env appmodulev2.Environment, cdc codec.BinaryCodec) *Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)

	k := &Keeper{
		authority:    authority,
		addressCodec: addressCodec, // TODO(@julienrbrt): Should we add address codec to the environment?
		environment:  env,
		params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.schema = schema

	return k
}
