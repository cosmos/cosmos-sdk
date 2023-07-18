package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ ViewKeeper = (*BaseViewKeeper)(nil)

// ViewKeeper defines a module interface that facilitates read only access to
// account balances.
type ViewKeeper interface {
	AppendLockedCoinsGetter(getter types.GetLockedCoinsFn)
	PrependLockedCoinsGetter(getter types.GetLockedCoinsFn)
	ClearLockedCoinsGetter()

	ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool

	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetAccountsBalances(ctx sdk.Context) []types.Balance
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	UnvestedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
}

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	ak       types.AccountKeeper

	lockedCoinsGetter *lockedCoinsGetter
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, ak types.AccountKeeper) BaseViewKeeper {
	rv := BaseViewKeeper{
		cdc:               cdc,
		storeKey:          storeKey,
		ak:                ak,
		lockedCoinsGetter: newLockedCoinsGetter(),
	}
	rv.AppendLockedCoinsGetter(rv.UnvestedCoins)
	return rv
}

// AppendLockedCoinsGetter adds the provided GetLockedCoinsFn to run after previously provided getters.
// The provided getter is wrapped in another that prevents it from returning zero and negative coin amounts.
func (k BaseViewKeeper) AppendLockedCoinsGetter(getter types.GetLockedCoinsFn) {
	k.lockedCoinsGetter.append(getLockedCoinsFnWrapper(getter))
}

// PrependLockedCoinsGetter adds the provided GetLockedCoinsFn to run before previously provided getters.
// The provided getter is wrapped in another that prevents it from returning zero and negative coin amounts.
func (k BaseViewKeeper) PrependLockedCoinsGetter(getter types.GetLockedCoinsFn) {
	k.lockedCoinsGetter.prepend(getLockedCoinsFnWrapper(getter))
}

// ClearLockedCoinsGetter removes the locked coins getter (if there is one).
func (k BaseViewKeeper) ClearLockedCoinsGetter() {
	k.lockedCoinsGetter.clear()
}

// Logger returns a module-specific logger.
func (k BaseViewKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
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

// GetAccountsBalances returns all the accounts balances from the store.
func (k BaseViewKeeper) GetAccountsBalances(ctx sdk.Context) []types.Balance {
	balances := make([]types.Balance, 0)
	mapAddressToBalancesIdx := make(map[string]int)

	k.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
		idx, ok := mapAddressToBalancesIdx[addr.String()]
		if ok {
			// address is already on the set of accounts balances
			balances[idx].Coins = balances[idx].Coins.Add(balance)
			balances[idx].Coins.Sort()
			return false
		}

		accountBalance := types.Balance{
			Address: addr.String(),
			Coins:   sdk.NewCoins(balance),
		}
		balances = append(balances, accountBalance)
		mapAddressToBalancesIdx[addr.String()] = len(balances) - 1
		return false
	})

	return balances
}

// GetBalance returns the balance of a specific denomination for a given account
// by address.
func (k BaseViewKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	accountStore := k.getAccountStore(ctx, addr)
	bz := accountStore.Get([]byte(denom))
	balance, err := UnmarshalBalanceCompat(k.cdc, bz, denom)
	if err != nil {
		panic(err)
	}

	return balance
}

// IterateAccountBalances iterates over the balances of a single account and
// provides the token balance to a callback. If true is returned from the
// callback, iteration is halted.
func (k BaseViewKeeper) IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool) {
	accountStore := k.getAccountStore(ctx, addr)

	iterator := accountStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key())
		balance, err := UnmarshalBalanceCompat(k.cdc, iterator.Value(), denom)
		if err != nil {
			panic(err)
		}

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
		address, denom, err := types.AddressAndDenomFromBalancesStore(iterator.Key())
		if err != nil {
			k.Logger(ctx).With("key", iterator.Key(), "err", err).Error("failed to get address from balances store")
			// TODO: revisit, for now, panic here to keep same behavior as in 0.42
			// ref: https://github.com/cosmos/cosmos-sdk/issues/7409
			panic(err)
		}

		balance, err := UnmarshalBalanceCompat(k.cdc, iterator.Value(), denom)
		if err != nil {
			panic(err)
		}

		if cb(address, balance) {
			break
		}
	}
}

// LockedCoins returns all the coins that are not spendable (i.e. locked) for an account by address.
func (k BaseViewKeeper) LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.lockedCoinsGetter.getLockedCoins(ctx, addr)
}

// UnvestedCoins returns all the coins that are locked due to a vesting schedule.
// It is appended as a GetLockedCoinsFn during NewBaseViewKeeper.
//
// You probably want to call LockedCoins instead. This function is primarily made public
// so that, externally, it can be re-injected after a call to ClearLockedCoinsGetter.
func (k BaseViewKeeper) UnvestedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	acc := k.ak.GetAccount(ctx, addr)
	if acc != nil {
		vacc, ok := acc.(types.VestingAccount)
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
	total := k.GetAllBalances(ctx, addr)
	locked := k.LockedCoins(ctx, addr)

	unlocked, hasNeg := total.SafeSub(locked...)
	if !hasNeg {
		return unlocked
	}

	spendable := sdk.Coins{}
	for _, coin := range unlocked {
		if coin.IsPositive() {
			spendable = append(spendable, coin)
		}
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

	vacc, ok := acc.(types.VestingAccount)
	if ok {
		ogv := vacc.GetOriginalVesting()
		if ogv.IsAnyGT(balances) {
			return fmt.Errorf("vesting amount %s cannot be greater than total amount %s", ogv, balances)
		}
	}

	return nil
}

// getAccountStore gets the account store of the given address.
func (k BaseViewKeeper) getAccountStore(ctx sdk.Context, addr sdk.AccAddress) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.CreateAccountBalancesPrefix(addr))
}

// getDenomAddressPrefixStore returns a prefix store that acts as a reverse index
// between a denomination and account balance for that denomination.
func (k BaseViewKeeper) getDenomAddressPrefixStore(ctx sdk.Context, denom string) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateDenomAddressPrefix(denom))
}

// UnmarshalBalanceCompat unmarshal balance amount from storage, it's backward-compatible with the legacy format.
func UnmarshalBalanceCompat(cdc codec.BinaryCodec, bz []byte, denom string) (sdk.Coin, error) {
	amount := math.ZeroInt()
	if bz == nil {
		return sdk.NewCoin(denom, amount), nil
	}

	if err := amount.Unmarshal(bz); err != nil {
		// try to unmarshal with the legacy format.
		var balance sdk.Coin
		if cdc.Unmarshal(bz, &balance) != nil {
			// return with the original error
			return sdk.Coin{}, err
		}
		return balance, nil
	}

	return sdk.NewCoin(denom, amount), nil
}

// lockedCoinsGetter is a struct that houses a GetLockedCoinsFn.
// It exists so that the GetLockedCoinsFn can be updated in the ViewKeeper without needing to have a pointer receiver.
type lockedCoinsGetter struct {
	fn types.GetLockedCoinsFn
}

// newLockedCoinsGetter creates a new lockedCoinsGetter with nil getter.
func newLockedCoinsGetter() *lockedCoinsGetter {
	return &lockedCoinsGetter{
		fn: nil,
	}
}

// append adds the provided function to this, to be run after the existing function.
func (r *lockedCoinsGetter) append(fn types.GetLockedCoinsFn) {
	r.fn = r.fn.Then(fn)
}

// prepend adds the provided function to this, to be run before the existing function.
func (r *lockedCoinsGetter) prepend(restriction types.GetLockedCoinsFn) {
	r.fn = restriction.Then(r.fn)
}

// clear removes the GetLockedCoinsFn (sets it to nil).
func (r *lockedCoinsGetter) clear() {
	r.fn = nil
}

var _ types.GetLockedCoinsFn = lockedCoinsGetter{}.getLockedCoins

// getLockedCoins runs the GetLockedCoinsFn if there is one. If not, it's a no-op.
func (r lockedCoinsGetter) getLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if r.fn == nil {
		return sdk.NewCoins()
	}
	return r.fn(ctx, addr)
}

// getLockedCoinsFnWrapper returns a new GetLockedCoinsFn that calls the provided getter but ensures
// only positive coin entries are returned. Coin entries with zero or negative amounts are ignored.
func getLockedCoinsFnWrapper(getter types.GetLockedCoinsFn) types.GetLockedCoinsFn {
	return func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
		locked := getter(ctx, addr)
		rv := sdk.Coins{}
		for _, coin := range locked {
			if coin.IsPositive() {
				rv = rv.Add(coin)
			}
		}
		return rv
	}
}
