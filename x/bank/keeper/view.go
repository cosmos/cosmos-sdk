package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ ViewKeeper = (*BaseViewKeeper)(nil)

// ViewKeeper defines a module interface that facilitates read only access to
// account balances.
type ViewKeeper interface {
	ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool

	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
}

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	cdc      codec.Marshaler
	storeKey sdk.StoreKey
	ak       types.AccountKeeper
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(cdc codec.Marshaler, storeKey sdk.StoreKey, ak types.AccountKeeper) BaseViewKeeper {
	return BaseViewKeeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
	}
}

// Logger returns a module-specific logger.
func (k BaseViewKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HasBalance returns whether or not an account has at least amt balance.
func (k BaseViewKeeper) HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool {
	return k.GetBalance(ctx, addr, amt.Denom).IsGTE(amt)
}

// GetAllBalances returns all the account balances for the given account address.
func (k BaseViewKeeper) GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	balances := sdk.NewCoins()
	k.IterateAccountBalances(ctx, addr, func(balance sdk.Coin) bool {
		balances = balances.Add(balance)
		return false
	})

	return balances.Sort()
}

// GetBalance returns the balance of a specific denomination for a given account
// by address.
func (k BaseViewKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())

	bz := accountStore.Get([]byte(denom))
	if bz == nil {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}

	var balance sdk.Coin
	k.cdc.MustUnmarshalBinaryBare(bz, &balance)

	return balance
}

// IterateAccountBalances iterates over the balances of a single account and
// provides the token balance to a callback. If true is returned from the
// callback, iteration is halted.
func (k BaseViewKeeper) IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool) {
	store := ctx.KVStore(k.storeKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())

	iterator := accountStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var balance sdk.Coin
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &balance)

		if cb(balance) {
			break
		}
	}
}

// IterateAllBalances iterates over all the balances of all accounts and
// denominations that are provided to a callback. If true is returned from the
// callback, iteration is halted.
func (k BaseViewKeeper) IterateAllBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) {
	store := ctx.KVStore(k.storeKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)

	iterator := balancesStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address := types.AddressFromBalancesStore(iterator.Key())

		var balance sdk.Coin
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &balance)

		if cb(address, balance) {
			break
		}
	}
}

// LockedCoins returns all the coins that are not spendable (i.e. locked) for an
// account by address. For standard accounts, the result will always be no coins.
// For vesting accounts, LockedCoins is delegated to the concrete vesting account
// type.
func (k BaseViewKeeper) LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	acc := k.ak.GetAccount(ctx, addr)
	if acc != nil {
		vacc, ok := acc.(vestexported.VestingAccount)
		if ok {
			return vacc.LockedCoins(ctx.BlockTime())
		}
	}

	return sdk.NewCoins()
}

// SpendableCoins returns the total balances of spendable coins for an account
// by address. If the account has no spendable coins, an empty Coins slice is
// returned.
func (k BaseViewKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	balances := k.GetAllBalances(ctx, addr)
	locked := k.LockedCoins(ctx, addr)

	spendable, hasNeg := balances.SafeSub(locked)
	if hasNeg {
		return sdk.NewCoins()
	}

	return spendable
}

// ValidateBalance validates all balances for a given account address returning
// an error if any balance is invalid. It will check for vesting account types
// and validate the balances against the original vesting balances.
//
// CONTRACT: ValidateBalance should only be called upon genesis state. In the
// case of vesting accounts, balances may change in a valid manner that would
// otherwise yield an error from this call.
func (k BaseViewKeeper) ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	balances := k.GetAllBalances(ctx, addr)
	if !balances.IsValid() {
		return fmt.Errorf("account balance of %s is invalid", balances)
	}

	vacc, ok := acc.(vestexported.VestingAccount)
	if ok {
		ogv := vacc.GetOriginalVesting()
		if ogv.IsAnyGT(balances) {
			return fmt.Errorf("vesting amount %s cannot be greater than total amount %s", ogv, balances)
		}
	}

	return nil
}
