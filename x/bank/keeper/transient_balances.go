package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// getTransientAccountStore gets the transient account store of the given address.
func (k BaseViewKeeper) getTransientAccountStore(ctx sdk.Context, addr sdk.AccAddress) prefix.Store {
	store := ctx.TransientStore(k.tStoreKey)
	return prefix.NewStore(store, types.CreateAccountBalancesPrefix(addr))
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
	store := ctx.TransientStore(k.tStoreKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)

	iterator := balancesStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address, _, err := types.AddressAndDenomFromBalancesStore(iterator.Key())
		if err != nil {
			k.Logger(ctx).With("key", iterator.Key(), "err", err).Error("failed to get address from balances store")
			continue
		}

		var balance sdk.Coin
		k.cdc.MustUnmarshal(iterator.Value(), &balance)

		if cb(address, balance) {
			break
		}
	}
}
