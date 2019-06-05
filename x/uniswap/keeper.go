package uniswap

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/log"
)

var (
	fee sdk.Dec // TODO: replace with fee param
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
	fee, err := sdk.NewDecFromStr("0.003")
	if err != nil {
		panic("could not construct fee")
	}

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
	key := GetExchangeKey()
	bz := store.Get(key)
	if bz != nil {
		panic("exchange already exists")
	}
	store.Set(key, sdk.NewDec(0))
}

// Deposit adds the specified amount of UNI to the associated account
func (keeper Keeper) Deposit(ctx sdk.Context, amt sdk.Dec, addr sdk.AccAddress) {
	var balance sdk.Dec

	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	bz := store.Get(key)
	if bz != nil {
		balance = keeper.decodeBalance(bz)
	}

	balance.Add(amt)
	store.Set(key, keeper.encodeBalance(balance))

	return
}

// Withdraw removes the specified amount of UNI from the associated account
func (keeper Keeper) Withdraw(ctx sdk.Context, amt sdk.Dec, addr sdk.AccAddress) {
	var balance sdk.Dec

	store := ctx.KVStore(keeper.storeKey)
	key := GetUNIBalancesKey(addr)
	bz := store.Get(key)
	if bz != nil {
		balance = keeper.decodeBalance(bz)
	}

	balance.Sub(amt)
	store.Set(key, keeper.encodeBalance(balance))
}

// GetInputAmount returns the amount of coins sold (calculated) given the output amount being bought (exact)
// The fee is included in the output coins being bought
func (keeper Keeper) GetInputAmount(ctx sdk.Context, outputAmt sdk.Dec, inputDenom, outputDenom string) sdk.Dec {
	store := ctx.KVStore(keeper.storeKey)
	inputReserve := store.Get(GetExchangeKey(inputDenom))
	if inputReserve == nil {
		panic("exchange for input denomination does not exist")
	}
	outputReserve := store.Get(GetExchangeKey(outputDenom))
	if outputReserve == nil {
		panic("exchange for output denomination does not exist")
	}

	// TODO: verify
	feeAmt := outputAmt.Mul(fee)
	numerator := inputReserve.Mul(outputReserve)
	denominator := outputReserve.Sub(outputAmt.Add(feeAmt))
	return numerator/denominator + 1
}

// GetOutputAmount returns the amount of coins bought (calculated) given the output amount being sold (exact)
// The fee is included in the input coins being bought
func (keeper Keeper) GetOutputAmount(ctx sdk.Context, inputAmt sdk.Dec, inputDenom, outputDenom string) sdk.Dec {
	store := ctx.KVStore(keeper.storeKey)
	inputReserve := store.Get(GetExchangeKey(inputDenom))
	if inputReserve == nil {
		panic("exchange for input denomination does not exist")
	}
	outputReserve := store.Get(GetExchangeKey(outputDenom))
	if outputReserve == nil {
		panic("exchange for output denomination does not exist")
	}

	// TODO: verify
	feeAmt := inputAmt.Mul(fee)
	inputAmtWithFee := inputReserve.Add(feeAmt)
	numerator := inputAmtWithFee * outputReserve
	denominator := inputReserve + inputAmt.Sub(feeAmt)
	return numerator / denominator
}

// -----------------------------------------------------------------------------
// Misc.

func (keeper Keeper) decodeBalance(bz []byte) (balance sdk.Dec) {
	err := keeper.cdc.UnmarshalBinaryBare(bz, &balance)
	if err != nil {
		panic(err)
	}
	return
}

func (keeper Keeper) encodeBalance(balance sdk.Dec) (bz []byte) {
	bz, err := keeper.cdc.MarshalBinaryBare(balance)
	if err != nil {
		panic(err)
	}
	return
}
