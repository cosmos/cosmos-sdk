package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/params"
		"fmt"
	)

var (
	collectedFeesKey = []byte("collectedFees")
	feeTokenKey      = "fee/token"			// fee/token
	FeeThresholdKey  = "fee/threshold"			// fee/threshold
	FeeExchangeRatePrefix = "fee/exchange/rate/"	// fee/exchange/rate/<denomination>
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

// Clears the collected Fee Pool
func (fck FeeCollectionKeeper) ClearCollectedFees(ctx sdk.Context) {
	fck.setCollectedFees(ctx, sdk.Coins{})
}

func (fck FeeCollectionKeeper) FeePreprocess(ctx sdk.Context, coins sdk.Coins) sdk.Error {
	feeToken, err := fck.getter.GetString(ctx, feeTokenKey)
	if err != nil {
		panic(err)
	}
	feeThreshold, err := fck.getter.GetInt(ctx, FeeThresholdKey)
	if err != nil {
		panic(err)
	}

	equivalentTotalFee := sdk.ZeroRat()
	for _,coin := range coins {
		if coin.Denom != feeToken {
			exchangeRateKey := FeeExchangeRatePrefix + coin.Denom
			rateBytes := fck.getter.GetRaw(ctx, exchangeRateKey)
			if rateBytes == nil {
				continue
			}
			var exchangeRate sdk.Rat
			err := fck.cdc.UnmarshalBinary(rateBytes, &exchangeRate)
			if err != nil {
				panic(err)
			}
			equivalentFee := exchangeRate.Mul(sdk.NewRatFromInt(coin.Amount, sdk.OneInt()))
			equivalentTotalFee = equivalentTotalFee.Add(equivalentFee)

		} else {
			equivalentTotalFee = equivalentTotalFee.Add(sdk.NewRatFromInt(coin.Amount, sdk.OneInt()))
		}
	}

	if equivalentTotalFee.LT(sdk.NewRatFromInt(feeThreshold, sdk.OneInt())) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("equivalent total fee %s is less than threshold %s", equivalentTotalFee.String(), feeThreshold))
	}
	return nil
}

type GenesisState struct {
	FeeToken string `json:"fee_token"`
	Threshold int64 `json:"fee_threshold"`
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		FeeToken: "iGas",
		Threshold: 100,
	}
}

func InitGenesis(ctx sdk.Context, setter params.Setter, data GenesisState) {
	setter.SetString(ctx, feeTokenKey, data.FeeToken)
	setter.SetInt(ctx, FeeThresholdKey, sdk.NewInt(data.Threshold))
}