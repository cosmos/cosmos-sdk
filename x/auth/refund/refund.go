package refund

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func RefundFees(supplyKeeper types.SupplyKeeper, ctx sdk.Context, acc sdk.AccAddress, refundFees sdk.Coins) error {
	blockTime := ctx.BlockHeader().Time
	feeCollector := supplyKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	coins := feeCollector.GetCoins()

	if !refundFees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid refund fee amount: %s", refundFees)
	}

	// verify the account has enough funds to pay for fees
	_, hasNeg := coins.SafeSub(refundFees)
	if hasNeg {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds,
			"insufficient funds to refund for fees; %s < %s", coins, refundFees)
	}

	// Validate the account has enough "spendable" coins as this will cover cases
	// such as vesting accounts.
	spendableCoins := feeCollector.SpendableCoins(blockTime)
	if _, hasNeg := spendableCoins.SafeSub(refundFees); hasNeg {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds,
			"insufficient funds to pay for refund fees; %s < %s", spendableCoins, refundFees)
	}

	err := supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.FeeCollectorName, acc, refundFees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}