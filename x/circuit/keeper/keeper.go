package keeper

import (
	context "context"

	"github.com/cosmos/cosmos-sdk/codec"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/circuit/types"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService

	authority []byte

	addressCodec address.Codec

	Schema      collections.Schema
	Permissions collections.Map[[]byte, types.Permissions]
	DisableList collections.KeySet[string]
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, authority string, addressCodec address.Codec) Keeper {
	auth, err := addressCodec.StringToBytes(authority)
	if err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    auth,
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

func (k *Keeper) IsAllowed(ctx context.Context, msgURL string) (bool, error) {
	has, err := k.DisableList.Has(ctx, msgURL)
	return !has, err
}

func (k *Keeper) DisableMsg(ctx context.Context, msgURL string) error {
	return k.DisableList.Set(ctx, msgURL)
}

func (k *Keeper) EnableMsg(ctx context.Context, msgURL string) error {
	return k.DisableList.Remove(ctx, msgURL)
}
