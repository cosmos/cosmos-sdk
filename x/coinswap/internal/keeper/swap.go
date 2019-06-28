package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SwapCoins(ctx sdk.Context, coinSold, coinBought sdk.Coin) {
	if !k.HasCoins(ctx, msg.Sender, coinSold) {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("sender account does not have sufficient amount of %s to fulfill the swap order", coinSold.Denom)).Result()
	}

	moduleName := getModuleName(ctx, k, coinSold.Denom, coinBought.Denom)
	k.SendCoin(ctx, msg.Sender, moduleName, coinSold)
	k.RecieveCoin(ctx, msg.Sender, moduleName, coinBought)
	return calculatedAmount
}

// getInputAmount returns the amount of coins sold (calculated) given the output amount being bought (exact)
// The fee is included in the output coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
// TODO: replace FeeD and FeeN with updated formula using fee as sdk.Dec
func (k Keeper) GetInputAmount(ctx sdk.Context, outputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	inputReserve := k.GetReservePool(inputDenom)
	outputReserve := k.GetReservePool(outputDenom)
	params := k.GetFeeParams(ctx)

	numerator := inputReserve.Mul(outputReserve).Mul(params.FeeD)
	denominator := (outputReserve.Sub(outputAmt)).Mul(parans.FeeN)
	return numerator.Quo(denominator).Add(sdk.OneInt())
}

// getOutputAmount returns the amount of coins bought (calculated) given the input amount being sold (exact)
// The fee is included in the input coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
// TODO: replace FeeD and FeeN with updated formula using fee as sdk.Dec
func (k Keeper) GetOutputAmount(ctx sdk.Context, inputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	inputReserve := k.GetReservePool(inputDenom)
	outputReserve := k.GetReservePool(outputDenom)
	params := k.GetFeeParams(ctx)

	inputAmtWithFee := inputAmt.Mul(params.FeeN)
	numerator := inputAmtWithFee.Mul(outputReserve)
	denominator := inputReserve.Mul(params.FeeD).Add(inputAmtWithFee)
	return numerator.Quo(denominator)
}

// isDoubleSwap returns true if the trade requires a double swap.
func (k Keeper) IsDoubleSwap(ctx sdk.Context, denom1, denom2 string) bool {
	nativeDenom := k.GetNativeDenom(ctx)
	return denom1 == nativeDenom || denom2 == nativeDenom
}

// getModuleName returns the ModuleAccount name for the provided denominations.
// The trading pair names are always sorted alphabetically.
func (k Keeper) GetModuleName(ctx sdk.Context, denom1, denom2 string) string {

}
