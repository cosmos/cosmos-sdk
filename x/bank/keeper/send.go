package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)

	SendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	SendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error

	BlockedAddr(addr sdk.AccAddress) bool
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper

	cdc        codec.BinaryMarshaler
	ak         types.AccountKeeper
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace

	// list of addresses that are restricted from receiving transactions
	blockedAddrs map[string]bool
}

func NewBaseSendKeeper(
	cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, ak types.AccountKeeper, paramSpace paramtypes.Subspace, blockedAddrs map[string]bool,
) BaseSendKeeper {

	return BaseSendKeeper{
		BaseViewKeeper: NewBaseViewKeeper(cdc, storeKey, ak),
		cdc:            cdc,
		ak:             ak,
		storeKey:       storeKey,
		paramSpace:     paramSpace,
		blockedAddrs:   blockedAddrs,
	}
}

// GetParams returns the total set of bank parameters.
func (k BaseSendKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of bank parameters.
func (k BaseSendKeeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// InputOutputCoins performs multi-send functionality. It accepts a series of
// inputs that correspond to a series of outputs. It returns an error if the
// inputs and outputs don't lineup or if any single transfer of tokens fails.
func (k BaseSendKeeper) InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error {
	// Safety check ensuring that when sending coins the keeper must maintain the
	// Check supply invariant and validity of Coins.
	if err := types.ValidateInputsOutputs(inputs, outputs); err != nil {
		return err
	}

	for _, in := range inputs {
		inAddress, err := sdk.AccAddressFromBech32(in.Address)
		if err != nil {
			return err
		}

		err = k.subUnlockedCoins(ctx, inAddress, in.Coins)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeySender, in.Address),
			),
		)
	}

	for _, out := range outputs {
		outAddress, err := sdk.AccAddressFromBech32(out.Address)
		if err != nil {
			return err
		}
		err = k.increaseBalance(ctx, outAddress, out.Coins)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, out.Address),
				sdk.NewAttribute(sdk.AttributeKeyAmount, out.Coins.String()),
			),
		)

		// Create account if recipient does not exist.
		//
		// NOTE: This should ultimately be removed in favor a more flexible approach
		// such as delegated fee messages.
		acc := k.ak.GetAccount(ctx, outAddress)
		if acc == nil {
			defer telemetry.IncrCounter(1, "new", "account")
			k.ak.SetAccount(ctx, k.ak.NewAccountWithAddress(ctx, outAddress))
		}
	}

	return nil
}

// SendCoins transfers amt coins from a sending account to a receiving account.
// An error is returned upon failure.
func (k BaseSendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeyRecipient, toAddr.String()),
			sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
		),
	})

	err := k.subUnlockedCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	err = k.increaseBalance(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	// Create account if recipient does not exist.
	//
	// NOTE: This should ultimately be removed in favor a more flexible approach
	// such as delegated fee messages.
	acc := k.ak.GetAccount(ctx, toAddr)
	if acc == nil {
		defer telemetry.IncrCounter(1, "new", "account")
		k.ak.SetAccount(ctx, k.ak.NewAccountWithAddress(ctx, toAddr))
	}

	return nil
}

// subUnlockedCoins removes the unlocked amt coins of the given account. An error is
// returned if the resulting balance is negative or the initial amount is invalid.
func (k BaseSendKeeper) subUnlockedCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {

	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	unlockedCoins, currentBalance := k.spendableCoins(ctx, addr)
	coinsAfterSpend, ok := unlockedCoins.SafeSub(amt)
	if ok {
		return balanceError(k.GetAllBalances(ctx, addr), amt, coinsAfterSpend)
	}

	// set new balance
	newBalance := currentBalance.Sub(amt)

	err := k.setBalances(ctx, addr, newBalance)
	if err != nil {
		return err
	}

	// emit coin spent event
	ctx.EventManager().EmitEvent(
		types.NewCoinSpentEvent(addr, amt),
	)
	return nil
}

// increaseBalance increase the addr balance by the given amt. It emits a coin received event.
func (k BaseSendKeeper) increaseBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	if err := k.addCoins(ctx, addr, amt); err != nil {
		return err
	}

	// emit coin received event
	ctx.EventManager().EmitEvent(
		types.NewCoinReceivedEvent(addr, amt),
	)

	return nil
}

// addCoins increases addr balance by the provided amt. An error is returned if
// the initial amount is invalid or if any resulting new balance is negative. It does
// not emit any event because it is used by MintCoins and increaseBalance which need to emit
// two different events to properly support balance and supply tracking.
func (k BaseSendKeeper) addCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	for _, coin := range amt {
		balance := k.GetBalance(ctx, addr, coin.Denom)
		newBalance := balance.Add(coin)

		err := k.setBalance(ctx, addr, newBalance)
		if err != nil {
			return err
		}
	}
	return nil
}

// subCoins removes from addr balance the given amt. It does not do any check on coins that are
// locked or not, and it does not emit any event as it is used by decreaseBalance and BurnCoins
// which emit two distinguished events to properly support balance and supply tracking via events.
func (k BaseSendKeeper) subCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	balance := k.GetAllBalances(ctx, addr)

	balanceAfterSpend, ok := balance.SafeSub(amt)
	// if there are no negative coin balances set the new balance
	if ok {
		return k.setBalances(ctx, addr, balanceAfterSpend)
	}

	// find the negative balance
	for _, coin := range balanceAfterSpend {
		// skip positive balances
		if !coin.IsNegative() {
			continue
		}

		coinOwned := sdk.NewCoin(coin.Denom, balance.AmountOf(coin.Denom))
		coinToSpend := sdk.NewCoin(coin.Denom, amt.AmountOf(coin.Denom))

		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"%s is smaller than %s", coinOwned, coinToSpend,
		)
	}
	// we should not reach this point
	panic("sub coin operation failed due to negative balance, but no negative coin balance was found")
}

// clearBalances removes all balances for a given account by address.
func (k BaseSendKeeper) clearBalances(ctx sdk.Context, addr sdk.AccAddress) {
	keys := [][]byte{}
	k.IterateAccountBalances(ctx, addr, func(balance sdk.Coin) bool {
		keys = append(keys, []byte(balance.Denom))
		return false
	})

	accountStore := k.getAccountStore(ctx, addr)

	for _, key := range keys {
		accountStore.Delete(key)
	}
}

// setBalances sets the balance (multiple coins) for an account by address. It will
// clear out all balances prior to setting the new coins as to set existing balances
// to zero if they don't exist in amt. An error is returned upon failure.
func (k BaseSendKeeper) setBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error {
	k.clearBalances(ctx, addr)

	for _, balance := range balances {
		err := k.setBalance(ctx, addr, balance)
		if err != nil {
			return err
		}
	}

	return nil
}

// setBalance sets the coin balance for an account by address.
func (k BaseSendKeeper) setBalance(ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin) error {
	if !balance.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, balance.String())
	}

	accountStore := k.getAccountStore(ctx, addr)

	bz := k.cdc.MustMarshalBinaryBare(&balance)
	accountStore.Set([]byte(balance.Denom), bz)

	return nil
}

// SendEnabledCoins checks the coins provide and returns an ErrSendDisabled if
// any of the coins are not configured for sending.  Returns nil if sending is enabled
// for all provided coin
func (k BaseSendKeeper) SendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	for _, coin := range coins {
		if !k.SendEnabledCoin(ctx, coin) {
			return sdkerrors.Wrapf(types.ErrSendDisabled, "%s transfers are currently disabled", coin.Denom)
		}
	}
	return nil
}

// SendEnabledCoin returns the current SendEnabled status of the provided coin's denom
func (k BaseSendKeeper) SendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool {
	return k.GetParams(ctx).SendEnabledDenom(coin.Denom)
}

// BlockedAddr checks if a given address is restricted from
// receiving funds.
func (k BaseSendKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	return k.blockedAddrs[addr.String()]
}

func balanceError(ownedCoins, coinsToSpend, afterBalance sdk.Coins) error {
	for _, coin := range afterBalance {
		if !coin.IsNegative() {
			continue
		}

		ownedCoin := sdk.NewCoin(coin.Denom, ownedCoins.AmountOf(coin.Denom))
		spendCoin := sdk.NewCoin(coin.Denom, coinsToSpend.AmountOf(coin.Denom))

		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"%s is smaller than %s", ownedCoin, spendCoin,
		)
	}
	// we should never reach this point...
	return nil
}
