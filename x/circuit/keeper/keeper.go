package keeper

import (
	proto "github.com/cosmos/gogoproto/proto"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	storekey storetypes.StoreKey

	authority string
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		storekey:  storeKey,
		authority: authority,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) GetPermissions(ctx sdk.Context, address string) (*types.Permissions, error) {
	store := ctx.KVStore(k.storekey)

	key := types.CreateAddressPrefix(address)
	bz := store.Get(key)

	perms := &types.Permissions{}
	if err := proto.Unmarshal(bz, perms); err != nil {
		return &types.Permissions{}, err
	}

	return perms, nil
}

func (k *Keeper) SetPermissions(ctx sdk.Context, address string, perms *types.Permissions) error {
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
	return store.Has(types.CreateDisableMsgPrefix(msgURL))
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

		}
	}(iter)

	for ; iter.Valid(); iter.Next() {
		var perms types.Permissions
		err := proto.Unmarshal(iter.Value(), &perms)
		if err != nil {
			return
		}

		if cb(iter.Key()[len(types.AccountPermissionPrefix):], perms) {
			break
		}
	}
}

func (k *Keeper) IterateDisableLists(ctx sdk.Context, cb func(address []byte, perms types.Permissions) (stop bool)) {
	store := ctx.KVStore(k.storekey)

	iter := storetypes.KVStorePrefixIterator(store, types.DisableListPrefix)
	defer func(iter storetypes.Iterator) {
		err := iter.Close()
		if err != nil {

		}
	}(iter)

	for ; iter.Valid(); iter.Next() {
		var perms types.Permissions
		err := proto.Unmarshal(iter.Value(), &perms)
		if err != nil {
			return
		}

		if cb(iter.Key()[len(types.DisableListPrefix):], perms) {
			break
		}
	}
}
