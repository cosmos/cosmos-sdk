package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func AdjustGasPrice(oldPrice sdk.DecCoins, gasUsed uint64, params TierParams) sdk.DecCoins {
	// zero means don't adjust for block load
	if params.ChangeDenominator == 0 {
		return oldPrice
	}
	if params.ParentGasTarget == 0 {
		return oldPrice
	}

	// adjust the coins one by one
	newPrice := make(sdk.DecCoins, 0, len(oldPrice))
	for _, coin := range oldPrice {
		newAmount := AdjustPriceAmount(coin.Amount, gasUsed, &params)
		if newAmount.IsZero() {
			continue
		}
		newPrice = append(newPrice, sdk.NewDecCoinFromDec(coin.Denom, newAmount))
	}
	return newPrice
}

func AdjustPriceAmount(old sdk.Dec, gasUsed uint64, params *TierParams) sdk.Dec {
	if gasUsed == params.ParentGasTarget {
		return old
	} else if gasUsed > params.ParentGasTarget {
		gasUsedDelta := gasUsed - params.ParentGasTarget
		delta := old.Mul(
			sdk.NewDecFromUint64(gasUsedDelta),
		).Quo(
			sdk.NewDecFromUint64(params.ParentGasTarget),
		).Quo(
			sdk.NewDecFromUint64(uint64(params.ChangeDenominator)),
		)
		return old.Add(delta)
	} else {
		gasUsedDelta := params.ParentGasTarget - gasUsed
		delta := old.Mul(
			sdk.NewDecFromUint64(gasUsedDelta),
		).Quo(
			sdk.NewDecFromUint64(params.ParentGasTarget),
		).Quo(
			sdk.NewDecFromUint64(uint64(params.ChangeDenominator)),
		)
		return old.Sub(delta)
	}
}

func ToProtoGasPrices(prices []sdk.DecCoins) []*sdk.ProtoDecCoins {
	protoGasPrices := make([]*sdk.ProtoDecCoins, len(prices))
	for i, price := range prices {
		protoGasPrices[i] = &sdk.ProtoDecCoins{
			Coins: price,
		}
	}
	return protoGasPrices
}
