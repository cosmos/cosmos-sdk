package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
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

	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)

	SetBalance(ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin) error
	SetBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error

	GetSendEnabled(ctx sdk.Context) bool
	SetSendEnabled(ctx sdk.Context, enabled bool)

	BlacklistedAddr(addr sdk.AccAddress) bool
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper

	cdc        codec.Marshaler
	ak         types.AccountKeeper
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace

	// list of addresses that are restricted from receiving transactions
	blacklistedAddrs map[string]bool
}

func NewBaseSendKeeper(
	cdc codec.Marshaler, storeKey sdk.StoreKey, ak types.AccountKeeper, paramSpace paramtypes.Subspace, blacklistedAddrs map[string]bool,
) BaseSendKeeper {

	return BaseSendKeeper{
		BaseViewKeeper:   NewBaseViewKeeper(cdc, storeKey, ak),
		cdc:              cdc,
		ak:               ak,
		storeKey:         storeKey,
		paramSpace:       paramSpace,
		blacklistedAddrs: blacklistedAddrs,
	}
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
		_, err := k.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeySender, in.Address.String()),
			),
		)
	}

	for _, out := range outputs {
		_, err := k.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, out.Address.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, out.Coins.String()),
			),
		)
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
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
		),
	})

	_, err := k.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	_, err = k.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	// Create account if recipient does not exist.
	//
	// NOTE: This should ultimately be removed in favor a more flexible approach
	// such as delegated fee messages.
	acc := k.ak.GetAccount(ctx, toAddr)
	if acc == nil {
		k.ak.SetAccount(ctx, k.ak.NewAccountWithAddress(ctx, toAddr))
	}

	return nil
}

// SubtractCoins removes amt coins the account by the given address. An error is
// returned if the resulting balance is negative or the initial amount is invalid.
func (k BaseSendKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error) {
	if !amt.IsValid() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	resultCoins := sdk.NewCoins()
	lockedCoins := k.LockedCoins(ctx, addr)

	for _, coin := range amt {
		balance := k.GetBalance(ctx, addr, coin.Denom)
		locked := sdk.NewCoin(coin.Denom, lockedCoins.AmountOf(coin.Denom))
		spendable := balance.Sub(locked)

		_, hasNeg := sdk.Coins{spendable}.SafeSub(sdk.Coins{coin})
		if hasNeg {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is smaller than %s", spendable, coin)
		}

		newBalance := balance.Sub(coin)
		resultCoins = resultCoins.Add(newBalance)

		err := k.SetBalance(ctx, addr, newBalance)
		if err != nil {
			return nil, err
		}
	}

	return resultCoins, nil
}

// AddCoins adds amt to the account balance given by the provided address. An
// error is returned if the initial amount is invalid or if any resulting new
// balance is negative.
func (k BaseSendKeeper) AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error) {
	if !amt.IsValid() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	var resultCoins sdk.Coins

	for _, coin := range amt {
		balance := k.GetBalance(ctx, addr, coin.Denom)
		newBalance := balance.Add(coin)
		resultCoins = resultCoins.Add(newBalance)

		err := k.SetBalance(ctx, addr, newBalance)
		if err != nil {
			return nil, err
		}
	}

	return resultCoins, nil
}

// ClearBalances removes all balances for a given account by address.
func (k BaseSendKeeper) ClearBalances(ctx sdk.Context, addr sdk.AccAddress) {
	keys := [][]byte{}
	k.IterateAccountBalances(ctx, addr, func(balance sdk.Coin) bool {
		keys = append(keys, []byte(balance.Denom))
		return false
	})

	store := ctx.KVStore(k.storeKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())

	for _, key := range keys {
		accountStore.Delete(key)
	}
}

// SetBalances sets the balance (multiple coins) for an account by address. It will
// clear out all balances prior to setting the new coins as to set existing balances
// to zero if they don't exist in amt. An error is returned upon failure.
func (k BaseSendKeeper) SetBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error {
	k.ClearBalances(ctx, addr)

	for _, balance := range balances {
		err := k.SetBalance(ctx, addr, balance)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetBalance sets the coin balance for an account by address.
func (k BaseSendKeeper) SetBalance(ctx sdk.Context, addr sdk.AccAddress, balance sdk.Coin) error {
	if !balance.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, balance.String())
	}

	store := ctx.KVStore(k.storeKey)
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())

	bz := k.cdc.MustMarshalBinaryBare(&balance)
	accountStore.Set([]byte(balance.Denom), bz)

	return nil
}

// GetSendEnabled returns the current SendEnabled
func (k BaseSendKeeper) GetSendEnabled(ctx sdk.Context) bool {
	var enabled bool
	k.paramSpace.Get(ctx, types.ParamStoreKeySendEnabled, &enabled)
	return enabled
}

// SetSendEnabled sets the send enabled
func (k BaseSendKeeper) SetSendEnabled(ctx sdk.Context, enabled bool) {
	k.paramSpace.Set(ctx, types.ParamStoreKeySendEnabled, &enabled)
}

// BlacklistedAddr checks if a given address is blacklisted (i.e restricted from
// receiving funds)
func (k BaseSendKeeper) BlacklistedAddr(addr sdk.AccAddress) bool {
	return k.blacklistedAddrs[addr.String()]
}
