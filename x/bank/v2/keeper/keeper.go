package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/event"
	"cosmossdk.io/math"
	"cosmossdk.io/x/bank/v2/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Keeper defines the bank/v2 module keeper.
// All fields are not exported, as they should only be accessed through the module's.
type Keeper struct {
	appmodulev2.Environment

	ak           types.AccountKeeper
	authority    []byte
	addressCodec address.Codec
	schema       collections.Schema
	params       collections.Item[types.Params]
	balances     *collections.IndexedMap[collections.Pair[sdk.AccAddress, string], math.Int, BalancesIndexes]
	supply       collections.Map[string, math.Int]
}

func NewKeeper(authority []byte, addressCodec address.Codec, env appmodulev2.Environment, cdc codec.BinaryCodec, ak types.AccountKeeper) *Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)

	k := &Keeper{
		Environment:  env,
		ak:           ak,
		authority:    authority,
		addressCodec: addressCodec, // TODO(@julienrbrt): Should we add address codec to the environment?
		params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		balances:     collections.NewIndexedMap(sb, types.BalancesPrefix, "balances", collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), types.BalanceValueCodec, newBalancesIndexes(sb)),
		supply:       collections.NewMap(sb, types.SupplyKey, "supply", collections.StringKey, sdk.IntValue),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.schema = schema

	return k
}

// MintCoins creates new coins from thin air and adds it to the module account.
// An error is returned if the module account does not exist or is unauthorized.
func (k Keeper) MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error {
	
	// TODO: Mint restriction
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName)
	}

	if !acc.HasPermission(authtypes.Minter) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName)
	}

	if !amounts.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amounts.String())
	}

	err := k.addCoins(ctx, acc.GetAddress(), amounts)
	if err != nil {
		return err
	}

	for _, amount := range amounts {
		supply := k.GetSupply(ctx, amount.GetDenom())
		supply = supply.Add(amount)
		k.setSupply(ctx, supply)
	}

	k.Logger.Debug("minted coins from module account", "amount", amounts.String(), "from", moduleName)

	addrStr, err := k.addressCodec.BytesToString(acc.GetAddress())
	if err != nil {
		return err
	}

	// emit mint event
	return k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeCoinMint,
		event.NewAttribute(types.AttributeKeyMinter, addrStr),
		event.NewAttribute(sdk.AttributeKeyAmount, amounts.String()),
	)
}

// SendCoins transfers amt coins from a sending account to a receiving account.
// Function take sender & receipient as string.
// They can be sdk address or module name.
// An error is returned upon failure.
func (k Keeper) SendCoins(ctx context.Context, from, to string, amt sdk.Coins) error {
	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	var fromAddr, toAddr sdk.AccAddress
	var err error

	// Detect from & to is address format or module name
	fromAddr, err = sdk.AccAddressFromBech32(from)
	if err != nil {
		// Check if is a module name
		fromAddr = k.ak.GetModuleAddress(from)
		if fromAddr == nil {
			return fmt.Errorf("%s is not an address or module name", from)
		}
	}

	toAddr, err = sdk.AccAddressFromBech32(to)
	if err != nil {
		// Check if is a module name
		toAddr = k.ak.GetModuleAddress(to)
		if toAddr == nil {
			return fmt.Errorf("%s is not an address or module name", to)
		}
	}
	
	// TODO: Send restriction

	err = k.subUnlockedCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	err = k.addCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	fromAddrString, err := k.addressCodec.BytesToString(fromAddr)
	if err != nil {
		return err
	}
	toAddrString, err := k.addressCodec.BytesToString(toAddr)
	if err != nil {
		return err
	}

	return k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeTransfer,
		event.NewAttribute(types.AttributeKeyRecipient, toAddrString),
		event.NewAttribute(types.AttributeKeySender, fromAddrString),
		event.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
	)
}

// GetSupply retrieves the Supply from store
func (k Keeper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	amt, err := k.supply.Get(ctx, denom)
	if err != nil {
		return sdk.NewCoin(denom, math.ZeroInt())
	}
	return sdk.NewCoin(denom, amt)
}

// GetBalance returns the balance of a specific denomination for a given account
// by address.
func (k Keeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	amt, err := k.balances.Get(ctx, collections.Join(addr, denom))
	if err != nil {
		return sdk.NewCoin(denom, math.ZeroInt())
	}
	return sdk.NewCoin(denom, amt)
}

// subUnlockedCoins removes the unlocked amt coins of the given account.
// An error is returned if the resulting balance is negative.
//
// CONTRACT: The provided amount (amt) must be valid, non-negative coins.
//
// A coin_spent event is emitted after the operation.
func (k Keeper) subUnlockedCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		balance := k.GetBalance(ctx, addr, coin.Denom)
		spendable := sdk.Coins{balance}

		_, hasNeg := spendable.SafeSub(coin)
		if hasNeg {
			if len(spendable) == 0 {
				spendable = sdk.Coins{sdk.Coin{Denom: coin.Denom, Amount: math.ZeroInt()}}
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

	addrStr, err := k.addressCodec.BytesToString(addr)
	if err != nil {
		return err
	}

	return k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeCoinSpent,
		event.NewAttribute(types.AttributeKeySpender, addrStr),
		event.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
	)
}

// addCoins increases the balance of the given address by the specified amount.
//
// CONTRACT: The provided amount (amt) must be valid, non-negative coins.
//
// It emits a coin_received event after the operation.
func (k Keeper) addCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		balance := k.GetBalance(ctx, addr, coin.Denom)
		newBalance := balance.Add(coin)

		err := k.setBalance(ctx, addr, newBalance)
		if err != nil {
			return err
		}
	}

	addrStr, err := k.addressCodec.BytesToString(addr)
	if err != nil {
		return err
	}

	return k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeCoinReceived,
		event.NewAttribute(types.AttributeKeyReceiver, addrStr),
		event.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
	)
}

// setSupply sets the supply for the given coin
func (k Keeper) setSupply(ctx context.Context, coin sdk.Coin) {
	// Bank invariants and IBC requires to remove zero coins.
	if coin.IsZero() {
		_ = k.supply.Remove(ctx, coin.Denom)
	} else {
		_ = k.supply.Set(ctx, coin.Denom, coin.Amount)
	}
}

// setBalance sets the coin balance for an account by address.
func (k Keeper) setBalance(ctx context.Context, addr sdk.AccAddress, balance sdk.Coin) error {
	if !balance.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, balance.String())
	}

	// x/bank invariants prohibit persistence of zero balances
	if balance.IsZero() {
		err := k.balances.Remove(ctx, collections.Join(addr, balance.Denom))
		if err != nil {
			return err
		}
		return nil
	}
	return k.balances.Set(ctx, collections.Join(addr, balance.Denom), balance.Amount)
}

func newBalancesIndexes(sb *collections.SchemaBuilder) BalancesIndexes {
	return BalancesIndexes{
		Denom: indexes.NewReversePair[math.Int](
			sb, types.DenomAddressPrefix, "address_by_denom_index",
			collections.PairKeyCodec(sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), collections.StringKey), //nolint:staticcheck // Note: refer to the LengthPrefixedAddressKey docs to understand why we do this.
			indexes.WithReversePairUncheckedValue(),                                                          // denom to address indexes were stored as Key: Join(denom, address) Value: []byte{0}, this will migrate the value to []byte{} in a lazy way.
		),
	}
}

type BalancesIndexes struct {
	Denom *indexes.ReversePair[sdk.AccAddress, string, math.Int]
}
