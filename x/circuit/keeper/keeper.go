package keeper

import (
	proto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	storekey storetypes.StoreKey

	authority []byte

	addressCodec address.Codec
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(storeKey storetypes.StoreKey, authority string, addressCodec address.Codec) Keeper {
	auth, err := addressCodec.StringToBytes(authority)
	if err != nil {
		panic(err)
	}

	return Keeper{
		storekey:     storeKey,
		authority:    auth,
		addressCodec: addressCodec,
	}
}

func (k *Keeper) GetAuthority() []byte {
	return k.authority
}

func (k *Keeper) GetPermissions(ctx sdk.Context, address []byte) (*types.Permissions, error) {
	store := ctx.KVStore(k.storekey)

	key := types.CreateAddressPrefix(address)
	bz := store.Get(key)

	perms := &types.Permissions{}
	if err := proto.Unmarshal(bz, perms); err != nil {
		return &types.Permissions{}, err
	}

	return perms, nil
}

func (k *Keeper) SetPermissions(ctx sdk.Context, address []byte, perms *types.Permissions) error {
	store := ctx.KVStore(k.storekey)

	bz, err := proto.Marshal(perms)
	if err != nil {
		return err
	}

	key := types.CreateAddressPrefix(address)
	store.Set(key, bz)

	return nil
}

func (k *Keeper) IsAllowed(ctx sdk.Context, msgURL string) bool {
	store := ctx.KVStore(k.storekey)
	return !store.Has(types.CreateDisableMsgPrefix(msgURL))
}

func (k *Keeper) DisableMsg(ctx sdk.Context, msgURL string) {
	ctx.KVStore(k.storekey).Set(types.CreateDisableMsgPrefix(msgURL), []byte{})
}

func (k *Keeper) EnableMsg(ctx sdk.Context, msgURL string) {
	ctx.KVStore(k.storekey).Delete(types.CreateDisableMsgPrefix(msgURL))
}

func (k *Keeper) IteratePermissions(ctx sdk.Context, cb func(address []byte, perms types.Permissions) (stop bool)) {
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

		if cb(iter.Key()[len(types.AccountPermissionPrefix):], perms) {
			break
		}
	}
}

func (k *Keeper) IterateDisableLists(ctx sdk.Context, cb func(url []byte, perms types.Permissions) (stop bool)) {
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
