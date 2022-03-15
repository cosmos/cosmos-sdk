package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

// CheckTxFee implements the middleware.FeeMarket interface
func (k Keeper) CheckTxFee(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	// default to tier zero
	var feeTier uint32
	if hasExtOptsTx, ok := tx.(middleware.HasExtensionOptionsTx); ok {
		for _, opt := range hasExtOptsTx.GetExtensionOptions() {
			if tieredTxOpt, ok := opt.GetCachedValue().(*types.ExtensionOptionTieredTx); ok {
				feeTier = tieredTxOpt.FeeTier
				break
			}
		}
	}
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, fmt.Errorf("Tx must be a FeeTx")
	}
	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	params := k.GetParams(ctx)
	if int(feeTier) >= len(params.Tiers) {
		return nil, 0, fmt.Errorf("Invalid fee tier %d", feeTier)
	}
	tierParams := params.Tiers[feeTier]
	gasPrice, found := k.GetGasPrice(ctx, feeTier)
	if !found {
		gasPrice = tierParams.InitialGasPrice
	}
	var effectiveGasPrice sdk.Coins
	if !gasPrice.IsZero() {
		requiredFees := make(sdk.Coins, len(gasPrice))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = gasPrice * gasLimit.
		for i, gp := range gasPrice {
			fee := gp.Amount.Mul(sdk.NewDecFromUint64(gas))
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		coin := feeCoins.FirstGTECoin(requiredFees)
		if coin == nil {
			return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}
		effectiveGasPrice = sdk.NewCoins(sdk.NewCoin(coin.Denom, requiredFees.AmountOf(coin.Denom)))
	}
	return effectiveGasPrice, tierParams.Priority, nil
}
