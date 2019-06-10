package uniswap

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	if msg.IsBuyOrder {
		inputAmount := k.GetInputAmount(ctx, msg.SwapDenom, msg.Amount)

	} else {
		outputAmount := k.GetOutputAmount(ctx, msg.SwapDenom, msg.Amount)

	}

	return sdk.Result{}
}

// HandleMsgAddLiquidity handler for MsgAddLiquidity
// If the exchange does not exist, it will be created.
func HandleMsgAddLiquidity(ctx sdk.Context, msg MsgAddLiquidity, k Keeper) sdk.Result {
	// create exchange if it does not exist
	coinLiquidity, err := k.GetExchange(ctx, msg.ExchangeDenom)
	if err != nil {
		k.CreateExchange(ctx, msg.Denom)
	}

	nativeLiqudiity, err := k.GetExchange(ctx, NativeAsset)
	if err != nil {
		panic("error retrieving native asset total liquidity")
	}

	totalLiquidity, err := k.GetTotalLiquidity(ctx)
	if err != nil {
		panic("error retrieving total UNI liquidity")
	}

	MintedUNI := (totalLiquidity.Mul(msg.DepositAmount)).Quo(nativeLiquidity)
	coinDeposited := (totalLiquidity.Mul(msg.DepositAmount)).Quo(nativeLiquidity)
	nativeCoin := sdk.NewCoin(NativeAsset, msg.DepositAmount)
	exchangeCoin := sdk.NewCoin(msg.ExchangeDenom, coinDeposited)
	k.SendCoins(ctx, msg.Sender, exchangeAcc, nativeCoin)
	k.SendCoins(ctx, msg.Sender, exchangeAcc, exchangeCoin)
	k.Deposit(ctx, msg.Sender, UNIMinted)

	return sdk.Result{}
}

// HandleMsgRemoveLiquidity handler for MsgRemoveLiquidity
func HandleMsgRemoveLiquidity(ctx sdk.Context, msg MsgRemoveLiquidity, k Keeper) sdk.Result {
	// check if exchange exists
	totalLiquidity, err := k.GetExchange(ctx, msg.ExchangeDenom)
	if err != nil {
		return err
	}

	return sdk.Result{}
}

// GetInputAmount returns the amount of coins sold (calculated) given the output amount being bought (exact)
// The fee is included in the output coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdfhttps://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
func getInputAmount(ctx sdk.Context, k Keeper, outputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	inputReserve := k.GetExchange(inputDenom)
	outputReserve := k.GetExchange(outputDenom)
	params := k.GetFeeParams(ctx)

	numerator := inputReserve.Mul(outputReserve).Mul(params.FeeD)
	denominator := (outputReserve.Sub(outputAmt)).Mul(parans.FeeN)
	return numerator.Quo(denominator).Add(sdk.OneInt())
}

// GetOutputAmount returns the amount of coins bought (calculated) given the input amount being sold (exact)
// The fee is included in the input coins being bought
// https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf
func getOutputAmount(ctx sdk.Context, k Keeper, inputAmt sdk.Int, inputDenom, outputDenom string) sdk.Int {
	inputReserve := k.GetExchange(inputDenom)
	outputReserve := k.GetExchange(outputDenom)
	params := k.GetFeeParams(ctx)

	inputAmtWithFee := inputAmt.Mul(params.FeeN)
	numerator := inputAmtWithFee.Mul(outputReserve)
	denominator := inputReserve.Mul(params.FeeD).Add(inputAmtWithFee)
	return numerator.Quo(denominator)
}
