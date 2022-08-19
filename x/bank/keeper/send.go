package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	IsSendEnabledDenom(ctx sdk.Context, denom string) bool
	GetSendEnabledEntry(ctx sdk.Context, denom string) (types.SendEnabled, bool)
	SetSendEnabled(ctx sdk.Context, denom string, value bool)
	SetAllSendEnabled(ctx sdk.Context, sendEnableds []*types.SendEnabled)
	DeleteSendEnabled(ctx sdk.Context, denom string)
	IterateSendEnabledEntries(ctx sdk.Context, cb func(denom string, sendEnabled bool) (stop bool))
	GetAllSendEnabledEntries(ctx sdk.Context) []types.SendEnabled

	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error

	BlockedAddr(addr sdk.AccAddress) bool
	GetBlockedAddresses() map[string]bool

	GetAuthority() string
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper

	cdc      codec.BinaryCodec
	ak       types.AccountKeeper
	storeKey storetypes.StoreKey

	// list of addresses that are restricted from receiving transactions
	blockedAddrs map[string]bool

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewBaseSendKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak types.AccountKeeper,
	blockedAddrs map[string]bool,
	authority string,
) BaseSendKeeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid bank authority address: %w", err))
	}

	return BaseSendKeeper{
		BaseViewKeeper: NewBaseViewKeeper(cdc, storeKey, ak),
		cdc:            cdc,
		ak:             ak,
		storeKey:       storeKey,
		blockedAddrs:   blockedAddrs,
		authority:      authority,
	}
}

// GetAuthority returns the x/bank module's authority.
func (k BaseSendKeeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the total set of bank parameters.
func (k BaseSendKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the total set of bank parameters.
//
//nolint:staticcheck // params.SendEnabled is deprecated but it should be here regardless.
func (k BaseSendKeeper) SetParams(ctx sdk.Context, params types.Params) error {
	// normally SendEnabled is deprecated but we still support it for backwards compatibility
	// using params.Validate() would fail due to the SendEnabled deprecation
	if len(params.SendEnabled) > 0 {
		k.SetAllSendEnabled(ctx, params.SendEnabled)
		// override params without SendEnabled
		params = types.NewParams(params.DefaultSendEnabled)
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// InputOutputCoins performs multi-send functionality. It accepts a series of
// inputs that correspond to a series of outputs. It returns an error if the
// inputs and outputs don't line up or if any single transfer of tokens fails.
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
		err = k.addCoins(ctx, outAddress, out.Coins)
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
		accExists := k.ak.HasAccount(ctx, outAddress)
		if !accExists {
			defer telemetry.IncrCounter(1, "new", "account")
			k.ak.SetAccount(ctx, k.ak.NewAccountWithAddress(ctx, outAddress))
		}
	}

	return nil
}

// SendCoins transfers amt coins from a sending account to a receiving account.
// An error is returned upon failure.
func (k BaseSendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	err := k.subUnlockedCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	err = k.addCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	// Create account if recipient does not exist.
	//
	// NOTE: This should ultimately be removed in favor a more flexible approach
	// such as delegated fee messages.
	accExists := k.ak.HasAccount(ctx, toAddr)
	if !accExists {
		defer telemetry.IncrCounter(1, "new", "account")
		k.ak.SetAccount(ctx, k.ak.NewAccountWithAddress(ctx, toAddr))
	}

	// bech32 encoding is expensive! Only do it once for fromAddr
	fromAddrString := fromAddr.String()
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeyRecipient, toAddr.String()),
			sdk.NewAttribute(types.AttributeKeySender, fromAddrString),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeySender, fromAddrString),
		),
	})

	return nil
}

// subUnlockedCoins removes the unlocked amt coins of the given account. An error is
// returned if the resulting balance is negative or the initial amount is invalid.
// A coin_spent event is emitted after.
func (k BaseSendKeeper) subUnlockedCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	lockedCoins := k.LockedCoins(ctx, addr)

	for _, coin := range amt {
		balance := k.GetBalance(ctx, addr, coin.Denom)
		locked := sdk.NewCoin(coin.Denom, lockedCoins.AmountOf(coin.Denom))
		spendable := balance.Sub(locked)

		_, hasNeg := sdk.Coins{spendable}.SafeSub(coin)
		if hasNeg {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is smaller than %s", spendable, coin)
		}

		newBalance := balance.Sub(coin)

		err := k.setBalance(ctx, addr, newBalance)
		if err != nil {
			return err
		}
	}

	// emit coin spent event
	ctx.EventManager().EmitEvent(
		types.NewCoinSpentEvent(addr, amt),
	)
	return nil
}

// addCoins increase the addr balance by the given amt. Fails if the provided amt is invalid.
// It emits a coin received event.
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

	// emit coin received event
	ctx.EventManager().EmitEvent(
		types.NewCoinReceivedEvent(addr, amt),
	)

	return nil
}

// initBalances sets the balance (multiple coins) for an account by address.
// An error is returned upon failure.
func (k BaseSendKeeper) initBalances(ctx sdk.Context, addr sdk.AccAddress, balances sdk.Coins) error {
	accountStore := k.getAccountStore(ctx, addr)
	denomPrefixStores := make(map[string]prefix.Store) // memoize prefix stores

	for i := range balances {
		balance := balances[i]
		if !balance.IsValid() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, balance.String())
		}

		// x/bank invariants prohibit persistence of zero balances
		if !balance.IsZero() {
			amount, err := balance.Amount.Marshal()
			if err != nil {
				return err
			}
			accountStore.Set([]byte(balance.Denom), amount)

			denomPrefixStore, ok := denomPrefixStores[balance.Denom]
			if !ok {
				denomPrefixStore = k.getDenomAddressPrefixStore(ctx, balance.Denom)
				denomPrefixStores[balance.Denom] = denomPrefixStore
			}

			// Store a reverse index from denomination to account address with a
			// sentinel value.
			denomAddrKey := address.MustLengthPrefix(addr)
			if !denomPrefixStore.Has(denomAddrKey) {
				denomPrefixStore.Set(denomAddrKey, []byte{0})
			}
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
	denomPrefixStore := k.getDenomAddressPrefixStore(ctx, balance.Denom)

	// x/bank invariants prohibit persistence of zero balances
	if balance.IsZero() {
		accountStore.Delete([]byte(balance.Denom))
		denomPrefixStore.Delete(address.MustLengthPrefix(addr))
	} else {
		amount, err := balance.Amount.Marshal()
		if err != nil {
			return err
		}
		accountStore.Set([]byte(balance.Denom), amount)

		// Store a reverse index from denomination to account address with a
		// sentinel value.
		denomAddrKey := address.MustLengthPrefix(addr)
		if !denomPrefixStore.Has(denomAddrKey) {
			denomPrefixStore.Set(denomAddrKey, []byte{0})
		}
	}

	return nil
}

// IsSendEnabledCoins checks the coins provide and returns an ErrSendDisabled if
// any of the coins are not configured for sending.  Returns nil if sending is enabled
// for all provided coin
func (k BaseSendKeeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	if len(coins) == 0 {
		return nil
	}
	store := ctx.KVStore(k.storeKey)
	haveDefault := false
	var defaultVal bool
	getDefault := func() bool {
		if !haveDefault {
			defaultVal = k.GetParams(ctx).DefaultSendEnabled
			haveDefault = true
		}
		return defaultVal
	}
	for _, coin := range coins {
		if !k.getSendEnabledOrDefault(store, coin.Denom, getDefault) {
			return types.ErrSendDisabled.Wrapf("%s transfers are currently disabled", coin.Denom)
		}
	}
	return nil
}

// IsSendEnabledCoin returns the current SendEnabled status of the provided coin's denom
func (k BaseSendKeeper) IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool {
	return k.IsSendEnabledDenom(ctx, coin.Denom)
}

// BlockedAddr checks if a given address is restricted from
// receiving funds.
func (k BaseSendKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	return k.blockedAddrs[addr.String()]
}

// GetBlockedAddresses returns the full list of addresses restricted from receiving funds.
func (k BaseSendKeeper) GetBlockedAddresses() map[string]bool {
	return k.blockedAddrs
}

// IsSendEnabledDenom returns the current SendEnabled status of the provided denom.
func (k BaseSendKeeper) IsSendEnabledDenom(ctx sdk.Context, denom string) bool {
	return k.getSendEnabledOrDefault(ctx.KVStore(k.storeKey), denom, func() bool { return k.GetParams(ctx).DefaultSendEnabled })
}

// GetSendEnabledEntry gets a SendEnabled entry for the given denom.
// The second return argument is true iff a specific entry exists for the given denom.
func (k BaseSendKeeper) GetSendEnabledEntry(ctx sdk.Context, denom string) (types.SendEnabled, bool) {
	sendEnabled, found := k.getSendEnabled(ctx.KVStore(k.storeKey), denom)
	if !found {
		return types.SendEnabled{}, false
	}
	return types.SendEnabled{Denom: denom, Enabled: sendEnabled}, true
}

// SetSendEnabled sets the SendEnabled flag for a denom to the provided value.
func (k BaseSendKeeper) SetSendEnabled(ctx sdk.Context, denom string, value bool) {
	store := ctx.KVStore(k.storeKey)
	k.setSendEnabledEntry(store, denom, value)
}

// SetAllSendEnabled sets all the provided SendEnabled entries in the bank store.
func (k BaseSendKeeper) SetAllSendEnabled(ctx sdk.Context, sendEnableds []*types.SendEnabled) {
	store := ctx.KVStore(k.storeKey)
	for _, se := range sendEnableds {
		k.setSendEnabledEntry(store, se.Denom, se.Enabled)
	}
}

// setSendEnabledEntry sets SendEnabled for the given denom to the give value in the provided store.
func (k BaseSendKeeper) setSendEnabledEntry(store sdk.KVStore, denom string, value bool) {
	key := types.CreateSendEnabledKey(denom)
	val := types.ToBoolB(value)
	store.Set(key, []byte{val})
}

// DeleteSendEnabled deletes a SendEnabled flag for a denom.
func (k BaseSendKeeper) DeleteSendEnabled(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.CreateSendEnabledKey(denom))
}

// getSendEnabledPrefixStore gets a prefix store for the SendEnabled entries.
func (k BaseSendKeeper) getSendEnabledPrefixStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.SendEnabledPrefix)
}

// IterateSendEnabledEntries iterates over all the SendEnabled entries.
func (k BaseSendKeeper) IterateSendEnabledEntries(ctx sdk.Context, cb func(denom string, sendEnabled bool) bool) {
	seStore := k.getSendEnabledPrefixStore(ctx)

	iterator := seStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key())
		val := types.IsTrueB(iterator.Value())
		if cb(denom, val) {
			break
		}
	}
}

// GetAllSendEnabledEntries gets all the SendEnabled entries that are stored.
// Any denoms not returned use the default value (set in Params).
func (k BaseSendKeeper) GetAllSendEnabledEntries(ctx sdk.Context) []types.SendEnabled {
	var rv []types.SendEnabled
	k.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) bool {
		rv = append(rv, types.SendEnabled{Denom: denom, Enabled: sendEnabled})
		return false
	})
	return rv
}

// getSendEnabled returns whether send is enabled and whether that flag was set for a denom.
//
// Example usage:
//
//	store := ctx.KVStore(k.storeKey)
//	sendEnabled, found := getSendEnabled(store, "atom")
//	if !found {
//	    sendEnabled = DefaultSendEnabled
//	}
func (k BaseSendKeeper) getSendEnabled(store sdk.KVStore, denom string) (bool, bool) {
	key := types.CreateSendEnabledKey(denom)
	if !store.Has(key) {
		return false, false
	}
	v := store.Get(key)
	if len(v) != 1 {
		return false, false
	}
	switch v[0] {
	case types.TrueB:
		return true, true
	case types.FalseB:
		return false, true
	default:
		return false, false
	}
}

// getSendEnabledOrDefault gets the send_enabled value for a denom. If it's not in the store, this will return the result of the getDefault function.
func (k BaseSendKeeper) getSendEnabledOrDefault(store sdk.KVStore, denom string, getDefault func() bool) bool {
	sendEnabled, found := k.getSendEnabled(store, denom)
	if found {
		return sendEnabled
	}
	return getDefault()
}
