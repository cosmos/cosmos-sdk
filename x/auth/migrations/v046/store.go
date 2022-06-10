package v046

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/internal/conv"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func mapAccountAddressToAccountId(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AddressStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var acc types.AccountI
		if err := cdc.UnmarshalInterface(iterator.Value(), &acc); err != nil {
			return err
		}
		store.Set(types.AccountNumberStoreKey(acc.GetAccountNumber()), conv.UnsafeStrToBytes(acc.GetAddress().String()))
	}

	return nil
}

// MigrateStore performs in-place store migrations from v0.45 to v0.46. The
// migration includes:
// - Add an Account number as an index to get the account address
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	return mapAccountAddressToAccountId(ctx, storeKey, cdc)
}
