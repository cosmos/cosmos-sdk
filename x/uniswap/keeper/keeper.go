package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/log"
)

// fee = 1 - p
// p = feeN / feeD
var (
	// TODO: replace with fee param
	feeN sdk.Int // numerator for determining fee
	feeD sdk.Int // denominator for determining fee
)

type Keeper struct {
	// The key used to access the store TODO update
	storeKey sdk.StoreKey

	// The reference to the CoinKeeper to modify balances after swaps or liquidity is deposited/withdrawn
	ck BankKeeper

	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec
}

// NewKeeper returns a uniswap keeper. It handles:
// - creating new exchanges
// - facilitating swaps
// - users adding liquidity to exchanges
// - users removing liquidity to exchanges
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ck BankKeeper) Keeper {
	// TODO: replace with param
	feeN = sdk.NewInt(997)
	feeD = sdk.NewInt(100)
	return Keeper{
		storeKey: key,
		ck:       ck,
		cdc:      cdc,
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/uniswap")
}

// CreateExchange initializes a new exchange pair between the new coin and the native asset
func (keeper Keeper) CreateExchange(ctx sdk.Context, newCoinDenom string) {
	store := ctx.KVStore(keeper.storeKey)
	key := GetExchangeKey(newCoinDenom)
	bz := store.Get(key)
	if bz != nil {
		panic("exchange already exists")
	}

	store.Set(key, keeper.encode(sdk.NewInt(0)))
}

// Deposit adds the specified amount of UNI to the associated account
func (keeper Keeper) Deposit(ctx sdk.Context, amt sdk.Int, addr sdk.AccAddress) {
	var balance sdk.Int

	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	bz := store.Get(key)
	if bz != nil {
		balance = keeper.decode(bz)
	}

	balance.Add(amt)
	store.Set(key, keeper.encode(balance))

	return
}

// Withdraw removes the specified amount of UNI from the associated account
func (keeper Keeper) Withdraw(ctx sdk.Context, amt sdk.Int, addr sdk.AccAddress) {
	var balance sdk.Int

	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	bz := store.Get(key)
	if bz != nil {
		balance = keeper.decode(bz)
	}

	balance.Sub(amt)
	store.Set(key, keeper.encode(balance))
}

// GetInputAmount returns the amount of coins sold (calculated) given the output amount being bought (exact)
// The fee is included in the output coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdfhttps://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
func (keeper Keeper) GetInputAmount(ctx sdk.Context, outputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GetExchangeKey(inputDenom))
	if bz == nil {
		panic("exchange for input denomination does not exist")
	}
	inputReserve := keeper.decode(bz)
	bz = store.Get(GetExchangeKey(outputDenom))
	if bz == nil {
		panic("exchange for output denomination does not exist")
	}
	outputReserve := keeper.decode(bz)

	numerator := inputReserve.Mul(outputReserve).Mul(feeD)
	denominator := (outputReserve.Sub(outputAmt)).Mul(feeN)
	return numerator.Quo(denominator).Add(sdk.NewInt(1))
}

// GetOutputAmount returns the amount of coins bought (calculated) given the input amount being sold (exact)
// The fee is included in the input coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
func (keeper Keeper) GetOutputAmount(ctx sdk.Context, inputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(GetExchangeKey(inputDenom))
	if bz == nil {
		panic("exchange for input denomination does not exist")
	}
	inputReserve := keeper.decode(bz)
	bz = store.Get(GetExchangeKey(outputDenom))
	if bz == nil {
		panic("exchange for output denomination does not exist")
	}
	outputReserve := keeper.decode(bz)

	inputAmtWithFee := inputAmt.Mul(feeN)
	numerator := inputAmtWithFee.Mul(outputReserve)
	denominator := inputReserve.Mul(feeD).Add(inputAmtWithFee)
	return numerator.Quo(denominator)
}

// -----------------------------------------------------------------------------
// Misc.

func (keeper Keeper) decode(bz []byte) (balance sdk.Int) {
	err := keeper.cdc.UnmarshalBinaryBare(bz, &balance)
	if err != nil {
		panic(err)
	}
	return
}

func (keeper Keeper) encode(balance sdk.Int) (bz []byte) {
	bz, err := keeper.cdc.MarshalBinaryBare(balance)
	if err != nil {
		panic(err)
	}
	return
}
