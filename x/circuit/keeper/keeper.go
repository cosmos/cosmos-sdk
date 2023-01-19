package keeper

import (
	proto "github.com/cosmos/gogoproto/proto"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
)

// keeper definines the circuit module's keeper.
type Keeper struct {
	key storetypes.StoreKey

	authority string
}

// contructs a new Circuit Keeper instance
func NewKeeper(storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		key:       storeKey,
		authority: authority,
	}
}

func (k *Keeper) GetAuthority() string /* string */ {
	return k.authority
}

func (k *Keeper) GetPermissions(ctx sdk.Context, address []byte) (*types.Permissions, error) {
	store := ctx.KVStore(k.key)

	key := types.CreateAddressPrefix(address)
	bz := store.Get(key)

	perms := &types.Permissions{}
	if err := proto.Unmarshal(bz, perms); err != nil {
		return &types.Permissions{}, err
	}

	return perms, nil
}

func (k *Keeper) SetPermissions(ctx sdk.Context, address []byte, perms *types.Permissions) error {
	store := ctx.KVStore(k.key)

	bz, err := proto.Marshal(perms)
	if err != nil {
		return err
	}

	key := types.CreateAddressPrefix(address)
	store.Set(key, bz)

	return nil
}

func (k *Keeper) IsMsgDisabled(ctx sdk.Context, msgURL string) bool {
	store := ctx.KVStore(k.key)
	return store.Has(types.CreateDisableMsgPrefix(msgURL))
}

func (k *Keeper) DisableMsg(ctx sdk.Context, msgURL string) {
	ctx.KVStore(k.key).Set(types.CreateDisableMsgPrefix(msgURL), []byte{})
}

func (k *Keeper) EnableMsg(ctx sdk.Context, msgURL string) {
	ctx.KVStore(k.key).Delete(types.CreateDisableMsgPrefix(msgURL))
}

func (k *Keeper) IteratePermissions(ctx sdk.Context, cb func(address []byte, perms types.Permissions) (stop bool)) {
	store := ctx.KVStore(k.key)

	iter := sdk.KVStorePrefixIterator(store, types.AccountPermissionPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var perms types.Permissions
		proto.Unmarshal(iter.Value(), &perms)

		if cb(iter.Key()[len(types.AccountPermissionPrefix):], perms) {
			break
		}
	}
}

func (k *Keeper) IterateDisableLists(ctx sdk.Context, cb func(address []byte, perms types.Permissions) (stop bool)) {
	store := ctx.KVStore(k.key)

	iter := sdk.KVStorePrefixIterator(store, types.DisableListPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var perms types.Permissions
		proto.Unmarshal(iter.Value(), &perms)

		if cb(iter.Key()[len(types.DisableListPrefix):], perms) {
			break
		}
	}
}
