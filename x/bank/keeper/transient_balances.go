package keeper

import (
	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// getTransientAccountStore gets the transient account store of the given address.
func (k BaseViewKeeper) getTransientAccountStore(ctx sdk.Context, addr sdk.AccAddress) prefix.Store {
	store := k.tStoreService.OpenTransientStore(ctx)
	return prefix.NewStore(runtime.KVStoreAdapter(store), createAccountBalancesPrefix(addr))
}

// setTransientBalance sets the transient coin balance for an account by address.
func (k BaseSendKeeper) setTransientBalance(ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin) {
	accountStore := k.getTransientAccountStore(ctx, addr)

	bz := k.cdc.MustMarshal(&balance)
	accountStore.Set([]byte(balance.Denom), bz)
}

func (k BaseKeeper) EmitAllTransientBalances(ctx sdk.Context) {
	balanceUpdates := k.GetAllTransientAccountBalanceUpdates(ctx)
	if len(balanceUpdates) > 0 {
		ctx.EventManager().EmitTypedEvent(&types.EventSetBalances{
			BalanceUpdates: balanceUpdates,
		})
	}
}

// GetAllTransientAccountBalanceUpdates returns all the transient accounts balances from the transient store.
func (k BaseViewKeeper) GetAllTransientAccountBalanceUpdates(ctx sdk.Context) []*types.BalanceUpdate {
	balanceUpdates := make([]*types.BalanceUpdate, 0)

	k.IterateAllTransientBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
		balanceUpdate := &types.BalanceUpdate{
			Addr:  addr.Bytes(),
			Denom: []byte(balance.Denom),
			Amt:   balance.Amount,
		}
		balanceUpdates = append(balanceUpdates, balanceUpdate)
		return false
	})

	return balanceUpdates
}

// IterateAllTransientBalances iterates over all transient balances of all accounts and
// denominations that are provided to a callback. If true is returned from the
// callback, iteration is halted.
func (k BaseViewKeeper) IterateAllTransientBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) {
	store := k.tStoreService.OpenTransientStore(ctx)
	balancesStore := prefix.NewStore(runtime.KVStoreAdapter(store), types.BalancesPrefix)

	iterator := balancesStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address, _, err := addressAndDenomFromBalancesStore(iterator.Key())
		if err != nil {
			k.Logger().With("key", iterator.Key(), "err", err).Error("failed to get address from balances store")
			continue
		}

		var balance sdk.Coin
		k.cdc.MustUnmarshal(iterator.Value(), &balance)

		if cb(address, balance) {
			break
		}
	}
}

func createAccountBalancesPrefix(addr []byte) []byte {
	return append(types.BalancesPrefix.Bytes(), address.MustLengthPrefix(addr)...)
}

func addressAndDenomFromBalancesStore(key []byte) (sdk.AccAddress, string, error) {
	if len(key) == 0 {
		return nil, "", types.ErrInvalidKey
	}

	kv.AssertKeyAtLeastLength(key, 1)

	addrBound := int(key[0])

	if len(key)-1 < addrBound {
		return nil, "", types.ErrInvalidKey
	}

	return key[1 : addrBound+1], string(key[addrBound+1:]), nil
}
