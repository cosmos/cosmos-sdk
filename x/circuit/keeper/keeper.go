package keeper

import (
	context "context"

	proto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit/types"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	storeService store.KVStoreService

	authority []byte

	addressCodec address.Codec
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(storeService store.KVStoreService, authority string, addressCodec address.Codec) Keeper {
	auth, err := addressCodec.StringToBytes(authority)
	if err != nil {
		panic(err)
	}

	return Keeper{
		storeService: storeService,
		authority:    auth,
		addressCodec: addressCodec,
	}
}

func (k *Keeper) GetAuthority() []byte {
	return k.authority
}

func (k *Keeper) GetPermissions(ctx context.Context, address []byte) (*types.Permissions, error) {
	store := k.storeService.OpenKVStore(ctx)

	key := types.CreateAddressPrefix(address)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}

	perms := &types.Permissions{}
	if err := proto.Unmarshal(bz, perms); err != nil {
		return &types.Permissions{}, err
	}

	return perms, nil
}

func (k *Keeper) SetPermissions(ctx context.Context, address []byte, perms *types.Permissions) error {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := proto.Marshal(perms)
	if err != nil {
		return err
	}

	key := types.CreateAddressPrefix(address)
	return store.Set(key, bz)
}

func (k *Keeper) IsAllowed(ctx context.Context, msgURL string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.CreateDisableMsgPrefix(msgURL))
	return !has, err
}

func (k *Keeper) DisableMsg(ctx context.Context, msgURL string) error {
	return k.storeService.OpenKVStore(ctx).Set(types.CreateDisableMsgPrefix(msgURL), []byte{})
}

func (k *Keeper) EnableMsg(ctx context.Context, msgURL string) error {
	return k.storeService.OpenKVStore(ctx).Delete(types.CreateDisableMsgPrefix(msgURL))
}

func (k *Keeper) IteratePermissions(ctx context.Context, cb func(address []byte, perms types.Permissions) (stop bool)) error {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.AccountPermissionPrefix, storetypes.PrefixEndBytes(types.AccountPermissionPrefix))
	if err != nil {
		return err
	}

	defer func(iter storetypes.Iterator) {
		err := iter.Close()
		if err != nil {
			return
		}
	}(iter)

	for ; iter.Valid(); iter.Next() {
		var perms types.Permissions
		err := proto.Unmarshal(iter.Value(), &perms)
		if err != nil {
			panic(err)
		}

		if cb(iter.Key()[len(types.AccountPermissionPrefix):], perms) {
			break
		}
	}
}

func (k *Keeper) IterateDisableLists(ctx context.Context, cb func(url []byte, perms types.Permissions) (stop bool)) {
	store := ctx.KVStore(k.storekey)

	iter := storetypes.KVStorePrefixIterator(store, types.AccountPermissionPrefix)
	defer func(iter storetypes.Iterator) {
		err := iter.Close()
		if err != nil {
			return
		}
	}(iter)

	for ; iter.Valid(); iter.Next() {
		var perms types.Permissions
		err := proto.Unmarshal(iter.Value(), &perms)
		if err != nil {
			panic(err)
		}

		if cb(iter.Key()[len(types.DisableListPrefix):], perms) {
			break
		}
	}
}
