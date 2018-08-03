package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/params"
	"fmt"
)

var (
	collectedFeesKey = []byte("collectedFees")
	NativeFeeTokenKey = "feeToken/native"
	NativeGasPriceThresholdKey  = "feeToken/native/gasPrice/threshold"
	FeeExchangeRatePrefix = "feeToken/derived/exchange/rate/"	//  key = feeToken/derived/exchange/rate/<denomination>, rate = BigInt(value)/10^18
	Precision = int64(1000000000000000000) //10^18
)

// This FeeCollectionKeeper handles collection of fees in the anteHandler
// and setting of MinFees for different fee tokens
type FeeCollectionKeeper struct {

	getter params.Getter

	// The (unexposed) key used to access the fee store from the Context.
	key sdk.StoreKey

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewFeeKeeper returns a new FeeKeeper
func NewFeeCollectionKeeper(cdc *wire.Codec, key sdk.StoreKey, getter params.Getter) FeeCollectionKeeper {
	return FeeCollectionKeeper{
		key: key,
		cdc: cdc,
		getter: getter,
	}
}

// Adds to Collected Fee Pool
func (fck FeeCollectionKeeper) GetCollectedFees(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(fck.key)
	bz := store.Get(collectedFeesKey)
	if bz == nil {
		return sdk.Coins{}
	}

	feePool := &(sdk.Coins{})
	fck.cdc.MustUnmarshalBinary(bz, feePool)
	return *feePool
}

// Sets to Collected Fee Pool
func (fck FeeCollectionKeeper) setCollectedFees(ctx sdk.Context, coins sdk.Coins) {
	bz := fck.cdc.MustMarshalBinary(coins)
	store := ctx.KVStore(fck.key)
	store.Set(collectedFeesKey, bz)
}

// Adds to Collected Fee Pool
func (fck FeeCollectionKeeper) addCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	newCoins := fck.GetCollectedFees(ctx).Plus(coins)
	fck.setCollectedFees(ctx, newCoins)

	return newCoins
}

func (fck FeeCollectionKeeper) refundCollectedFees(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	newCoins := fck.GetCollectedFees(ctx).Minus(coins)
	if !newCoins.IsNotNegative() {
		panic("fee collector contains negative coins")
	}
	fck.setCollectedFees(ctx, newCoins)

	return newCoins
}


// Clears the collected Fee Pool
func (fck FeeCollectionKeeper) ClearCollectedFees(ctx sdk.Context) {
	fck.setCollectedFees(ctx, sdk.Coins{})
}

func (fck FeeCollectionKeeper) FeePreprocess(ctx sdk.Context, coins sdk.Coins, gasLimit int64) sdk.Error {
	if gasLimit <=0 {
		return sdk.ErrInternal(fmt.Sprintf("gaslimit %s should be larger than 0", gasLimit))
	}
	nativeFeeToken, err := fck.getter.GetString(ctx, NativeFeeTokenKey)
	if err != nil {
		panic(err)
	}
	nativeGasPriceThreshold, err := fck.getter.GetInt(ctx, NativeGasPriceThresholdKey)
	if err != nil {
		panic(err)
	}


	equivalentTotalFee := sdk.NewInt(0)
	for _,coin := range coins {
		if coin.Denom != nativeFeeToken {
			exchangeRateKey := FeeExchangeRatePrefix + coin.Denom
			rate, err := fck.getter.GetInt(ctx, exchangeRateKey)
			if err != nil {
				panic(err)
			}

			equivalentFee := coin.Amount.Mul(rate).Div(sdk.NewInt(Precision))
			equivalentTotalFee = equivalentTotalFee.Add(equivalentFee)

		} else {
			equivalentTotalFee = equivalentTotalFee.Add(coin.Amount)
		}
	}

	gasPrice := equivalentTotalFee.Div(sdk.NewInt(gasLimit))
	if gasPrice.LT(nativeGasPriceThreshold) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("gas price %s is less than threshold %s", gasPrice.String(), nativeGasPriceThreshold.String()))
	}
	return nil
}

type GenesisState struct {
	FeeTokenNative string `json:"fee_token_native"`
	GasPriceThreshold int64 `json:"gas_price_threshold"`
}

func DefaultGenesisStateForTest() GenesisState {
	return GenesisState{
		FeeTokenNative: "atom",
		GasPriceThreshold: 0,
	}
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		FeeTokenNative: "atom",
		GasPriceThreshold: 5,
	}
}

func InitGenesis(ctx sdk.Context, setter params.Setter, data GenesisState) {
	setter.SetString(ctx, NativeFeeTokenKey, data.FeeTokenNative)
	setter.SetInt(ctx, NativeGasPriceThresholdKey, sdk.NewInt(data.GasPriceThreshold))
}