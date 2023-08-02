package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/math"
)

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	AppendSendRestriction(restriction types.SendRestrictionFn)
	PrependSendRestriction(restriction types.SendRestrictionFn)
	ClearSendRestriction()

	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error

	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)

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

	GetAuthority() string
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper

	cdc        codec.BinaryCodec
	ak         types.AccountKeeper
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	// list of addresses that are restricted from receiving transactions
	blockedAddrs map[string]bool

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	sendRestriction *sendRestriction
}

func NewBaseSendKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak types.AccountKeeper,
	paramSpace paramtypes.Subspace,
	blockedAddrs map[string]bool,
	authority string,
) BaseSendKeeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid bank authority address: %w", err))
	}

	return BaseSendKeeper{
		BaseViewKeeper:  NewBaseViewKeeper(cdc, storeKey, ak),
		cdc:             cdc,
		ak:              ak,
		storeKey:        storeKey,
		paramSpace:      paramSpace,
		blockedAddrs:    blockedAddrs,
		authority:       authority,
		sendRestriction: newSendRestriction(),
	}
}

// AppendSendRestriction adds the provided SendRestrictionFn to run after previously provided restrictions.
func (k BaseSendKeeper) AppendSendRestriction(restriction types.SendRestrictionFn) {
	k.sendRestriction.append(restriction)
}

// PrependSendRestriction adds the provided SendRestrictionFn to run before previously provided restrictions.
func (k BaseSendKeeper) PrependSendRestriction(restriction types.SendRestrictionFn) {
	k.sendRestriction.prepend(restriction)
}

// ClearSendRestriction removes the send restriction (if there is one).
func (k BaseSendKeeper) ClearSendRestriction() {
	k.sendRestriction.clear()
}

func (k BaseSendKeeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the total set of bank parameters.
func (k BaseSendKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of bank parameters.
func (k BaseSendKeeper) SetParams(ctx sdk.Context, params types.Params) {
	if len(params.SendEnabled) > 0 {
		k.SetAllSendEnabled(ctx, params.SendEnabled)
	}
	p := types.NewParams(params.DefaultSendEnabled)
	k.paramSpace.SetParamSet(ctx, &p)
}

// InputOutputCoins performs multi-send functionality. There must be exactly one input and/or exactly one output.
// It returns an error if the input and outputs don't line up or if any single transfer of tokens fails.
func (k BaseSendKeeper) InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error {
	// As a keeper function, we can't assume that MsgMultiSend.ValidateBasic was called on these inputs and outputs.
	if err := types.ValidateInputsOutputs(inputs, outputs); err != nil {
		return err
	}

	// Remove the funds from the inputs first as that's the most common point of failure.
	for i, input := range inputs {
		inAddress := sdk.MustAccAddressFromBech32(input.Address)
		err := k.subUnlockedCoins(ctx, inAddress, input.Coins)
		if err != nil {
			return fmt.Errorf("input[%d] %s: %w", i, input.Address, err)
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeySender, input.Address),
			),
		)
	}

	// Create a map of AccAddress (cast to string) to the amount that that address will get.
	// The keys are the addresses that come back from the send restriction, not necessarily the addresses in the outputs.
	// Keep track of the order of the output address too since looping over a map is non-deterministic.
	toOutput := make(map[string]sdk.Coins)
	outputOrder := make([]sdk.AccAddress, 0, math.MaxInt(len(inputs), len(outputs)))
	// applySendRestriction will make the call to the send restriction function,
	// and update the toOutput and outputOrder values accordingly.
	applySendRestriction := func(inputAddress, outputAddress string, coins sdk.Coins) error {
		inAddr := sdk.MustAccAddressFromBech32(inputAddress)
		outAddrOrig := sdk.MustAccAddressFromBech32(outputAddress)
		outAddr, err := k.sendRestriction.apply(ctx, inAddr, outAddrOrig, coins)
		if err != nil {
			return err
		}
		amt, known := toOutput[string(outAddr)]
		if !known {
			outputOrder = append(outputOrder, outAddr)
		}
		toOutput[string(outAddr)] = amt.Add(coins...)
		return nil
	}

	// If there's multiple inputs, we apply the send restriction for each input.
	// Otherwise, apply the send restriction for each output.
	// ValidateInputsOutputs prevents many-to-many.
	if len(inputs) > 1 {
		for _, input := range inputs {
			err := applySendRestriction(input.Address, outputs[0].Address, input.Coins)
			if err != nil {
				return err
			}
		}
	} else {
		for _, output := range outputs {
			err := applySendRestriction(inputs[0].Address, output.Address, output.Coins)
			if err != nil {
				return err
			}
		}
	}

	// Finally, add the coins to the appropriate account(s).
	for _, outAddress := range outputOrder {
		amt := toOutput[string(outAddress)]
		err := k.addCoins(ctx, outAddress, amt)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, outAddress.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
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
// If the receiving account is quarantined, and not set to auto-accept funds from the sender,
// the coins will be transferred from the fromAddr to the quarantine funds holder account and be recorded as quarantined.
// Otherwise, the coins will be transferred from the fromAddr to the toAddr.
// An error is returned upon failure.
func (k BaseSendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	var err error
	err = k.subUnlockedCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	toAddr, err = k.sendRestriction.apply(ctx, fromAddr, toAddr, amt)
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

		spendable, hasNeg := sdk.Coins{balance}.SafeSub(locked)
		if hasNeg {
			return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
				"locked amount exceeds account balance funds: %s > %s", locked, balance)
		}

		if _, hasNeg = spendable.SafeSub(coin); hasNeg {
			// If spendable is zero, .String() would just be "". So give it a zero coin entry for that message.
			if spendable.IsZero() {
				spendable = sdk.Coins{sdk.Coin{Denom: coin.Denom, Amount: sdk.ZeroInt()}}
			}
			return errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"spendable balance %s is smaller than %s",
				spendable, coin,
			)
		}

		newBalance := balance.Sub(coin)

		if err := k.setBalance(ctx, addr, newBalance); err != nil {
			return err
		}
	}

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

// sendRestriction is a struct that houses a SendRestrictionFn.
// It exists so that the SendRestrictionFn can be updated in the SendKeeper without needing to have a pointer receiver.
type sendRestriction struct {
	fn types.SendRestrictionFn
}

// newSendRestriction creates a new sendRestriction with nil send restriction.
func newSendRestriction() *sendRestriction {
	return &sendRestriction{
		fn: nil,
	}
}

// append adds the provided restriction to this, to be run after the existing function.
func (r *sendRestriction) append(restriction types.SendRestrictionFn) {
	r.fn = r.fn.Then(restriction)
}

// prepend adds the provided restriction to this, to be run before the existing function.
func (r *sendRestriction) prepend(restriction types.SendRestrictionFn) {
	r.fn = restriction.Then(r.fn)
}

// clear removes the send restriction (sets it to nil).
func (r *sendRestriction) clear() {
	r.fn = nil
}

var _ types.SendRestrictionFn = sendRestriction{}.apply

// apply applies the send restriction if there is one. If not, it's a no-op.
func (r sendRestriction) apply(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
	if r.fn == nil {
		return toAddr, nil
	}
	return r.fn(ctx, fromAddr, toAddr, amt)
}
