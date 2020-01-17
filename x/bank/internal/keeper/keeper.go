package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var _ Keeper = (*BaseKeeper)(nil)

var balancesPrefix = []byte("balances")

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper

	DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
}

// BaseKeeper manages transfers between accounts. It implements the Keeper interface.
type BaseKeeper struct {
	BaseSendKeeper

	storeKey sdk.StoreKey

	ak         types.AccountKeeper
	paramSpace params.Subspace
}

// NewBaseKeeper returns a new BaseKeeper
func NewBaseKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper, paramSpace params.Subspace, blacklistedAddrs map[string]bool,
) BaseKeeper {

	ps := paramSpace.WithKeyTable(types.ParamKeyTable())
	return BaseKeeper{
		BaseSendKeeper: NewBaseSendKeeper(cdc, storeKey, ak, ps, blacklistedAddrs),
		ak:             ak,
		paramSpace:     ps,
	}
}

// DelegateCoins performs delegation by deducting amt coins from an account with
// address addr. For vesting accounts, delegations amounts are tracked for both
// vesting and vested coins.
// The coins are then transferred from the delegator address to a ModuleAccount address.
// If any of the delegation amounts are negative, an error is returned.
func (keeper BaseKeeper) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := keeper.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	for _, coin := range amt {
		// Need to manually do this as SubtractCoins will check for unspendable coins
		oldBalance := keeper.GetBalance(ctx, delegatorAddr, coin.Denom)
		if oldBalance.IsLT(coin) {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "insufficient account funds; %s < %s", oldBalance, amt,
			)
		}
		keeper.SetBalance(ctx, delegatorAddr, oldBalance.Sub(coin))
	}

	if err := keeper.trackDelegation(ctx, delegatorAddr, ctx.BlockHeader().Time, amt); err != nil {
		return sdkerrors.Wrap(err, "failed to track delegation")
	}

	_, err := keeper.AddCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// UndelegateCoins performs undelegation by crediting amt coins to an account with
// address addr. For vesting accounts, undelegation amounts are tracked for both
// vesting and vested coins.
// The coins are then transferred from a ModuleAccount address to the delegator address.
// If any of the undelegation amounts are negative, an error is returned.
// CONTRACT:  ModuleAccAddr is not for a vestion account
func (keeper BaseKeeper) UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := keeper.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	// Safe to use Substract coins because a moduleAcc shouldn't be vesting
	_, err := keeper.SubtractCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		return err
	}

	if err := keeper.trackUndelegation(ctx, delegatorAddr, amt); err != nil {
		return sdkerrors.Wrap(err, "failed to track undelegation")
	}

	_, err = keeper.AddCoins(ctx, delegatorAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error)

	SetBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) error
	SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error

	GetSendEnabled(ctx sdk.Context) bool
	SetSendEnabled(ctx sdk.Context, enabled bool)

	BlacklistedAddr(addr sdk.AccAddress) bool
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper

	cdc *codec.Codec

	ak         types.AccountKeeper
	storeKey   sdk.StoreKey
	paramSpace params.Subspace

	// list of addresses that are restricted from receiving transactions
	blacklistedAddrs map[string]bool
}

// NewBaseSendKeeper returns a new BaseSendKeeper.
func NewBaseSendKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper, paramSpace params.Subspace, blacklistedAddrs map[string]bool,
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

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseSendKeeper) InputOutputCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error {
	// Safety check ensuring that when sending coins the keeper must maintain the
	// Check supply invariant and validity of Coins.
	if err := types.ValidateInputsOutputs(inputs, outputs); err != nil {
		return err
	}

	for _, in := range inputs {
		_, err := keeper.SubtractCoins(ctx, in.Address, in.Coins)
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
		_, err := keeper.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTransfer,
				sdk.NewAttribute(types.AttributeKeyRecipient, out.Address.String()),
			),
		)
	}

	return nil
}

// SendCoins moves coins from one account to another
func (keeper BaseSendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
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

	_, err := keeper.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	_, err = keeper.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// SubtractCoins subtracts amt from the coins at the addr.
//
// CONTRACT: If the account is a vesting account, the amount has to be spendable.
func (keeper BaseSendKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error) {
	if !amt.IsValid() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	resultCoins := sdk.NewCoins()

	unspendableCoins := keeper.GetUnspendableCoins(ctx, addr)

	for _, coin := range amt {
		oldBalance := keeper.GetBalance(ctx, addr, coin.Denom)
		newBalance := oldBalance.Sub(coin)
		if newBalance.Amount.LT(unspendableCoins.AmountOf(coin.Denom)) {
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "insufficient account funds; resulting balance %s is below spendable limit", newBalance,
			)
		}
		keeper.SetBalance(ctx, addr, newBalance)
		resultCoins = resultCoins.Add(newBalance)
	}

	return resultCoins, nil
}

// AddCoins adds amt to the coins at the addr.
func (keeper BaseSendKeeper) AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, error) {
	if !amt.IsValid() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	var resultCoins sdk.Coins

	for _, coin := range amt {
		oldBalance := keeper.GetBalance(ctx, addr, coin.Denom)
		newBalance := oldBalance.Add(coin)
		if newBalance.IsNegative() {
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "insufficient account funds; resulting balance %s is negative", newBalance,
			)
		}
		keeper.SetBalance(ctx, addr, newBalance)
		resultCoins = resultCoins.Add(newBalance)
	}

	return resultCoins, nil
}

// SetCoins sets the balances for multiple denoms at the addr.
func (keeper BaseSendKeeper) SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		err := keeper.SetBalance(ctx, addr, coin)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetBalance sets the coin balance for a specific denom at the addr.
func (keeper BaseSendKeeper) SetBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) error {
	if !amt.IsValid() {
		sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	store := ctx.KVStore(keeper.storeKey)
	balancesStore := prefix.NewStore(store, balancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())

	bz := keeper.cdc.MustMarshalBinaryBare(amt)
	accountStore.Set([]byte(amt.Denom), bz)

	return nil
}

// GetSendEnabled returns the current SendEnabled
func (keeper BaseSendKeeper) GetSendEnabled(ctx sdk.Context) bool {
	var enabled bool
	keeper.paramSpace.Get(ctx, types.ParamStoreKeySendEnabled, &enabled)
	return enabled
}

// SetSendEnabled sets the send enabled
func (keeper BaseSendKeeper) SetSendEnabled(ctx sdk.Context, enabled bool) {
	keeper.paramSpace.Set(ctx, types.ParamStoreKeySendEnabled, &enabled)
}

// BlacklistedAddr checks if a given address is blacklisted (i.e restricted from
// receiving funds)
func (keeper BaseSendKeeper) BlacklistedAddr(addr sdk.AccAddress) bool {
	return keeper.blacklistedAddrs[addr.String()]
}

var _ ViewKeeper = (*BaseViewKeeper)(nil)

// ViewKeeper defines a module interface that facilitates read only access to
// account balances.
type ViewKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool

	GetUnspendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
}

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	ak       types.AccountKeeper
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding of balances.
	cdc *codec.Codec
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper) BaseViewKeeper {
	return BaseViewKeeper{
		cdc:      cdc,
		storeKey: storeKey,
		ak:       ak,
	}
}

// Logger returns a module-specific logger.
func (keeper BaseViewKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetBalance returns the balance of a specific denom at the addr.
func (keeper BaseViewKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(keeper.storeKey)
	balancesStore := prefix.NewStore(store, balancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())

	bz := accountStore.Get([]byte(denom))
	if bz == nil {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}
	var balance sdk.Coin
	keeper.cdc.MustUnmarshalBinaryBare(bz, &balance)
	return balance
}

// HasBalance returns whether or not an account has at least amt coins.
func (keeper BaseViewKeeper) HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool {
	return keeper.GetBalance(ctx, addr, amt.Denom).IsGTE(amt)
}

// IterateAccountBalances iterates over the balances of a single account and performs a callback function
func (keeper BaseViewKeeper) IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	balancesStore := prefix.NewStore(store, balancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr.Bytes())
	iterator := accountStore.Iterator(nil, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var coin sdk.Coin
		keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &coin)

		if cb(coin) {
			break
		}
	}
}

// IterateAllBalances iterates over all the balances of all accounts and denoms and performs a callback function
func (keeper BaseViewKeeper) IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	balancesStore := prefix.NewStore(store, balancesPrefix)
	iterator := balancesStore.Iterator(nil, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		address := decodeAddressFromBalanceKey(iterator.Key())
		var coin sdk.Coin
		keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &coin)

		if cb(address, coin) {
			break
		}
	}
}

func decodeAddressFromBalanceKey(key []byte) sdk.AccAddress {
	return sdk.AccAddress(key[len(balancesPrefix) : len(balancesPrefix)+sdk.AddrLen])
}

// GetUnspendableCoins returns a list of unspendable coins due to vesting
func (keeper BaseViewKeeper) GetUnspendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	acc := keeper.ak.GetAccount(ctx, addr)
	if acc != nil {
		vacc, ok := acc.(vestexported.VestingAccount)
		if ok {
			vestingDenoms := vacc.GetVestingCoins(ctx.BlockTime())
			currVestingDenomBalances := sdk.NewCoins()
			for _, denom := range vestingDenoms {
				currDenomBalance := keeper.GetBalance(ctx, acc.GetAddress(), denom.Denom)
				currVestingDenomBalances.Add(currDenomBalance)
			}
			return vacc.UnspendableCoins(currVestingDenomBalances, ctx.BlockTime())
		}
	}
	return sdk.NewCoins()
}

// CONTRACT: assumes that amt is valid.
func (keeper BaseKeeper) trackDelegation(ctx sdk.Context, addr sdk.AccAddress, blockTime time.Time, amt sdk.Coins) error {
	acc := keeper.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}
	vacc, ok := acc.(vestexported.VestingAccount)
	if ok {
		// TODO: return error on account.TrackDelegation
		vacc.TrackDelegation(blockTime, amt)
	}
	return nil
}

// CONTRACT: assumes that amt is valid.
func (keeper BaseKeeper) trackUndelegation(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	acc := keeper.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}
	vacc, ok := acc.(vestexported.VestingAccount)
	if ok {
		// TODO: return error on account.TrackUndelegation
		vacc.TrackUndelegation(amt)
	}
	return nil
}
