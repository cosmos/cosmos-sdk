package uniswap

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// NewHandler routes the messages to the handlers
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSwapOrder:
			return HandleMsgSwapOrder(ctx, msg, k)
		case MsgAddLiquidity:
			return HandleMsgAddLiquidity(ctx, msg, k)
		case MsgRemoveLiquidity:
			return HandleMsgRemoveLiquidity(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("unrecognized uniswap message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// HandleMsgSwapOrder handler for MsgSwapOrder
func HandleMsgSwapOrder(ctx sdk.Context, msg MsgSwapOrder, k Keeper) sdk.Result {
	var caclulatedAmount sdk.Int

	// check that deadline has not passed
	if msg.ctx.BlockTime.After(msg.Deadline) {
		return ErrInvalidDeadline(DefaultCodespace, "deadline has passed for MsgSwapOrder")
	}

	if msg.IsBuyOrder {
		calculatedAmount := k.GetInputAmount(ctx, msg.SwapDenom, msg.Amount)
		// ensure the maximum amount sender is willing to sell is less than
		// the calculated amount
		if msg.Input.Amount.LT(calculatedAmount) {
			return types.ErrInvalidBound(DefaultCodespace, "maximum amount (%d) to be sold was exceeded (%d)", msg.Input.Amount, calculatedAmount).Result()
		}

		coinSold := sdk.NewCoins(sdk.NewCoin(msg.Input.Denom, calculatedAmount))
		if !s.bk.HasCoins(ctx, msg.Sender, coinSold) {
			return ErrInsufficientAmount(DefaultCodespace, "sender account does not have sufficient funds to fulfill the swap order").Result()
		}

		err := k.sk.SendCoinsFromAccountToModule(ctx, msg.Sender, ModuleName, coinSold)
		if err != nil {
			return err.Result()
		}

		err = k.sk.SendCoinsFromModuleToAccount(ctx, ModuleName, msg.Sender, sdk.NewCoins(msg.Output))
		if err != nil {
			return err.Result()
		}

	} else {
		outputAmount := k.GetOutputAmount(ctx, msg.SwapDenom, msg.Amount)
		// ensure the minimum amount the sender is willing to buy is greater than
		// the calculated amount
		if msg.Output.Amount.GT(outputAmount) {
			return sdk.ErrInvalidBound(DefaultCodespace, "minimum amount (%d) to be sold was not met (%d)", msg.Output.Amount, calculatedAmount).Result()
		}

		coinSold := sdk.NewCoins(msg.Input)
		if !s.bk.HasCoins(ctx, msg.Sender, coinSold) {
			return ErrInsifficientAmount(DefaultCodespace, "sender account does not have sufficient funds to fulfill the swap order").Result()
		}

		err := k.sk.SendCoinsFromAccountToModule(ctx, msg.Sender, ModuleName, sdk.NewCoins(msg.Input))
		if err != nil {
			return err.Result()
		}

		err = k.sk.SendCoinsFromModuleToAccount(ctx, ModuleName, msg.Sender, sdk.NewCoins(sdk.NewCoin(msg.Output.Denom, calculatedAmount)))
		if err != nil {
			return err.Result()
		}

	}

	return sdk.Result{}
}

// HandleMsgAddLiquidity handler for MsgAddLiquidity
// If the reserve pool does not exist, it will be created.
func HandleMsgAddLiquidity(ctx sdk.Context, msg MsgAddLiquidity, k Keeper) sdk.Result {
	// check that deadline has not passed
	if msg.ctx.BlockTime.After(msg.Deadline) {
		return ErrInvalidDeadline(DefaultCodespace, "deadline has passed for MsgAddLiquidity")
	}

	// create reserve pool if it does not exist
	coinLiquidity := k.GetReservePool(ctx, msg.Deposit.Denom)
	if err != nil {
		k.CreateReservePool(ctx, msg.Denom)
	}

	nativeLiqudiity, err := k.GetReservePool(ctx, NativeAsset)
	if err != nil {
		panic("error retrieving native asset total liquidity")
	}

	totalUNI, err := k.GetTotalUNI(ctx)
	if err != nil {
		panic("error retrieving total UNI")
	}

	// calculate amount of UNI to be minted for sender
	// and coin amount to be deposited
	MintedUNI := (totalUNI.Mul(msg.DepositAmount)).Quo(nativeLiquidity)
	coinAmountDeposited := (totalUNI.Mul(msg.DepositAmount)).Quo(nativeLiquidity)
	nativeCoinDeposited := sdk.NewCoin(NativeAsset, msg.DepositAmount)
	coinDeposited := sdk.NewCoin(msg.Deposit.Denom, coinAmountDeposited)

	coins := sdk.NewCoins(nativeCoinDeposited, coinDeposited)
	if !s.bk.HasCoins(ctx, msg.Sender, coins) {
		return ErrInsufficientCoins(DefaultCodespace, "sender does not have sufficient funds to add liquidity").Result()
	}

	// transfer deposited liquidity into uniswaps ModuleAccount
	err := k.sk.SendCoinsFromAccountToModule(ctx, msg.Sender, ModuleName, coins)
	if err != nil {
		return err.Result()
	}

	// set updated total UNI
	totalUNI = totalUNI.Add(MintedUNI)
	k.SetTotalUNI(ctx, totalUNI)

	// update senders account with minted UNI
	UNIBalance := k.GetUNIForAddress(ctx, msg.Sender)
	UNIBalance = UNIBalance.Add(MintedUNI)
	k.SetUNIForAddress(ctx, UNIBalance)

	return sdk.Result{}
}

// HandleMsgRemoveLiquidity handler for MsgRemoveLiquidity
func HandleMsgRemoveLiquidity(ctx sdk.Context, msg MsgRemoveLiquidity, k Keeper) sdk.Result {
	// check that deadline has not passed
	if msg.ctx.BlockTime.After(msg.Deadline) {
		return ErrInvalidDeadline(DefaultCodespace, "deadline has passed for MsgRemoveLiquidity")
	}

	// check if reserve pool exists
	coinLiquidity, err := k.GetReservePool(ctx, msg.Withdraw.Denom)
	if err != nil {
		panic(fmt.Sprintf("error retrieving total liquidity for denomination: %s", msg.Withdraw.Denom))
	}

	nativeLiquidity, err := k.GetReservePool(ctx, NativeAsset)
	if err != nil {
		panic("error retrieving native asset total liquidity")
	}

	totalUNI, err := k.GetTotalUNI(ctx)
	if err != nil {
		panic("error retrieving total UNI")
	}

	// calculate amount of UNI to be burned for sender
	// and coin amount to be returned
	nativeWithdrawn := msg.WithdrawAmount.Mul(nativeLiquidity).Quo(totalUNI)
	coinWithdrawn := msg.WithdrawAmount.Mul(coinLiqudity).Quo(totalUNI)
	nativeCoin := sdk.NewCoin(nativeDenom, nativeWithdrawn)
	exchangeCoin = sdk.NewCoin(msg.Withdraw.Denom, coinWithdrawn)

	// transfer withdrawn liquidity from uniswaps ModuleAccount to sender's account
	err := k.sk.SendCoinsFromModuleToAccount(ctx, msg.Sender, ModuleName, sdk.NewCoins(nativeCoin, coinDeposited))
	if err != nil {
		return err.Result()
	}

	// set updated total UNI
	totalUNI = totalUNI.Add(MintedUNI)
	k.SetTotalUNI(ctx, totalUNI)

	// update senders account with minted UNI
	UNIBalance := k.GetUNIForAddress(ctx, msg.Sender)
	UNIBalance = UNIBalance.Add(MintedUNI)
	k.SetUNIForAddress(ctx, UNIBalance)

	return sdk.Result{}
}

// GetInputAmount returns the amount of coins sold (calculated) given the output amount being bought (exact)
// The fee is included in the output coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdfhttps://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
func getInputAmount(ctx sdk.Context, k Keeper, outputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	inputReserve := k.GetReservePool(inputDenom)
	outputReserve := k.GetReservePool(outputDenom)
	params := k.GetFeeParams(ctx)

	numerator := inputReserve.Mul(outputReserve).Mul(params.FeeD)
	denominator := (outputReserve.Sub(outputAmt)).Mul(parans.FeeN)
	return numerator.Quo(denominator).Add(sdk.OneInt())
}

// GetOutputAmount returns the amount of coins bought (calculated) given the input amount being sold (exact)
// The fee is included in the input coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
func getOutputAmount(ctx sdk.Context, k Keeper, inputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	inputReserve := k.GetReservePool(inputDenom)
	outputReserve := k.GetReservePool(outputDenom)
	params := k.GetFeeParams(ctx)

	inputAmtWithFee := inputAmt.Mul(params.FeeN)
	numerator := inputAmtWithFee.Mul(outputReserve)
	denominator := inputReserve.Mul(params.FeeD).Add(inputAmtWithFee)
	return numerator.Quo(denominator)
}
