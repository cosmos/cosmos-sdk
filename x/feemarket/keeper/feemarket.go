package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateFeeMarket updates the base fee and learning rate based on the
// AIMD learning rate adjustment algorithm. Note that if the fee market
// is disabled, this function will return without updating the fee market.
// This is executed in EndBlock which allows the next block's base fee to
// be readily available for wallets to estimate gas prices.
func (k *Keeper) UpdateFeeMarket(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(
		"updated the fee market",
		"params", params,
	)

	if !params.Enabled {
		return nil
	}

	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}

	// Update the learning rate based on the block utilization seen in the
	// current block. This is the AIMD learning rate adjustment algorithm.
	newLR := state.UpdateLearningRate(
		params,
	)

	// Update the base gas price based with the new learning rate and delta adjustment.
	newBaseGasPrice := state.UpdateBaseGasPrice(params)

	k.Logger(ctx).Info(
		"updated the fee market",
		"height", ctx.BlockHeight(),
		"new_base_gas_price", newBaseGasPrice,
		"new_learning_rate", newLR,
		"average_block_utilization", state.GetAverageUtilization(params),
		"net_block_utilization", state.GetNetUtilization(params),
	)

	// Increment the height of the state and set the new state.
	state.IncrementHeight()
	return k.SetState(ctx, state)
}

// GetBaseGasPrice returns the base fee from the fee market state.
func (k *Keeper) GetBaseGasPrice(ctx sdk.Context) (math.LegacyDec, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}

	return state.BaseGasPrice, nil
}

// GetLearningRate returns the learning rate from the fee market state.
func (k *Keeper) GetLearningRate(ctx sdk.Context) (math.LegacyDec, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}

	return state.LearningRate, nil
}

// GetMinGasPrice returns the mininum gas prices for given denom as sdk.DecCoins from the fee market state.
func (k *Keeper) GetMinGasPrice(ctx sdk.Context, denom string) (sdk.DecCoin, error) {
	baseGasPrice, err := k.GetBaseGasPrice(ctx)
	if err != nil {
		return sdk.DecCoin{}, err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return sdk.DecCoin{}, err
	}

	var gasPrice sdk.DecCoin

	if params.FeeDenom == denom {
		gasPrice = sdk.NewDecCoinFromDec(params.FeeDenom, baseGasPrice)
	} else {
		gasPrice, err = k.ResolveToDenom(ctx, sdk.NewDecCoinFromDec(params.FeeDenom, baseGasPrice), denom)
		if err != nil {
			return sdk.DecCoin{}, err
		}
	}

	return gasPrice, nil
}

// GetMinGasPrices returns the mininum gas prices as sdk.DecCoins from the fee market state.
func (k *Keeper) GetMinGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	baseGasPrice, err := k.GetBaseGasPrice(ctx)
	if err != nil {
		return sdk.NewDecCoins(), err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return sdk.NewDecCoins(), err
	}

	minGasPrice := sdk.NewDecCoinFromDec(params.FeeDenom, baseGasPrice)
	minGasPrices := sdk.NewDecCoins(minGasPrice)

	extraDenoms, err := k.resolver.ExtraDenoms(ctx)
	if err != nil {
		return sdk.NewDecCoins(), err
	}

	for _, denom := range extraDenoms {
		gasPrice, err := k.ResolveToDenom(ctx, minGasPrice, denom)
		if err != nil {
			k.Logger(ctx).Info(
				"failed to convert gas price",
				"min gas price", minGasPrice,
				"denom", denom,
			)
			continue
		}
		minGasPrices = minGasPrices.Add(gasPrice)
	}

	return minGasPrices, nil
}
