package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/bank/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Keeper defines the bank/v2 module keeper.
// All fields are not exported, as they should only be accessed through the module's.
type Keeper struct {
	appmodulev2.Environment

	authority    []byte
	addressCodec address.Codec
	schema       collections.Schema
	params       collections.Item[types.Params]
	balances     *collections.IndexedMap[collections.Pair[[]byte, string], math.Int, BalancesIndexes]
	supply       collections.Map[string, math.Int]
}

func NewKeeper(authority []byte, addressCodec address.Codec, env appmodulev2.Environment, cdc codec.BinaryCodec) *Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)

	k := &Keeper{
		Environment:  env,
		authority:    authority,
		addressCodec: addressCodec, // TODO(@julienrbrt): Should we add address codec to the environment?
		params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		balances:     collections.NewIndexedMap(sb, types.BalancesPrefix, "balances", collections.PairKeyCodec(collections.BytesKey, collections.StringKey), sdk.IntValue, newBalancesIndexes(sb)),
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
func (k Keeper) MintCoins(ctx context.Context, addr []byte, amounts sdk.Coins) error {
	// TODO: Mint restriction & permission

	if !amounts.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amounts.String())
	}

	err := k.addCoins(ctx, addr, amounts)
	if err != nil {
		return err
	}

	for _, amount := range amounts {
		supply := k.GetSupply(ctx, amount.GetDenom())
		supply = supply.Add(amount)
		k.setSupply(ctx, supply)
	}

	addrStr, err := k.addressCodec.BytesToString(addr)
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
// Function take sender & receipient as []byte.
// They can be sdk address or module name.
// An error is returned upon failure.
func (k Keeper) SendCoins(ctx context.Context, from, to []byte, amt sdk.Coins) error {
	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	var err error
	// TODO: Send restriction

	err = k.subUnlockedCoins(ctx, from, amt)
	if err != nil {
		return err
	}

	err = k.addCoins(ctx, to, amt)
	if err != nil {
		return err
	}

	fromAddrString, err := k.addressCodec.BytesToString(from)
	if err != nil {
		return err
	}
	toAddrString, err := k.addressCodec.BytesToString(to)
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
func (k Keeper) GetBalance(ctx context.Context, addr []byte, denom string) sdk.Coin {
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
func (k Keeper) subUnlockedCoins(ctx context.Context, addr []byte, amt sdk.Coins) error {
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
func (k Keeper) addCoins(ctx context.Context, addr []byte, amt sdk.Coins) error {
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
func (k Keeper) setBalance(ctx context.Context, addr []byte, balance sdk.Coin) error {
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
			collections.PairKeyCodec(collections.BytesKey, collections.StringKey),
			indexes.WithReversePairUncheckedValue(), // denom to address indexes were stored as Key: Join(denom, address) Value: []byte{0}, this will migrate the value to []byte{} in a lazy way.
		),
	}
}

type BalancesIndexes struct {
	Denom *indexes.ReversePair[[]byte, string, math.Int]
}
