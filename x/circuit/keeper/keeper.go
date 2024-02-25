package keeper

import (
	context "context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	cdc codec.BinaryCodec
	env appmodule.Environment

	authority []byte

	addressCodec address.Codec

	Schema collections.Schema
	// Permissions contains the permissions for each account
	Permissions collections.Map[[]byte, types.Permissions]
	// DisableList contains the message URLs that are disabled
	DisableList collections.KeySet[string]
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(env appmodule.Environment, cdc codec.BinaryCodec, authority string, addressCodec address.Codec) Keeper {
	auth, err := addressCodec.StringToBytes(authority)
	if err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(env.KVStoreService)

	k := Keeper{
		cdc:          cdc,
		authority:    auth,
		env:          env,
		addressCodec: addressCodec,
		Permissions: collections.NewMap(
			sb,
			types.AccountPermissionPrefix,
			"permissions",
			collections.BytesKey,
			codec.CollValue[types.Permissions](cdc),
		),
		DisableList: collections.NewKeySet(
			sb,
			types.DisableListPrefix,
			"disable_list",
			collections.StringKey,
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

func (k *Keeper) GetAuthority() []byte {
	return k.authority
}

// IsAllowed returns true when msg URL is not found in the DisableList for given context, else false.
func (k *Keeper) IsAllowed(ctx context.Context, msgURL string) (bool, error) {
	has, err := k.DisableList.Has(ctx, msgURL)
	return !has, err
}
