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
	if msg.IsBuyOrder {
		inputAmount := k.GetInputAmount(ctx, msg.SwapDenom, msg.Amount)
		coinSold = sdk.NewCoin(msg.ExchangeDenom, inputAmount)

		k.supplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Sender, ModuleName, sdk.NewCoins(coinSold))
		k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, ModuleName, msg.Sender, sdk.NewCoins(msg.Amount))
	} else {
		outputAmount := k.GetOutputAmount(ctx, msg.SwapDenom, msg.Amount)
		coinBought := sdk.NewCoin(msg.ExchangeDenom, outputAmount)

		k.supplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Sender, ModuleName, sdk.NewCoins(msg.Amount))
		k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, ModuleName, msg.Sender, sdk.NewCoins(coinBought))
	}

	return sdk.Result{}
}

// HandleMsgAddLiquidity handler for MsgAddLiquidity
// If the exchange does not exist, it will be created.
func HandleMsgAddLiquidity(ctx sdk.Context, msg MsgAddLiquidity, k Keeper) sdk.Result {
	// create exchange if it does not exist
	coinLiquidity := k.GetExchange(ctx, msg.ExchangeDenom)
	if err != nil {
		k.CreateExchange(ctx, msg.Denom)
	}

	nativeLiqudiity, err := k.GetExchange(ctx, NativeAsset)
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
	coinDeposited := (totalUNI.Mul(msg.DepositAmount)).Quo(nativeLiquidity)
	nativeCoin := sdk.NewCoin(NativeAsset, msg.DepositAmount)
	exchangeCoin := sdk.NewCoin(msg.ExchangeDenom, coinDeposited)

	// transfer deposited liquidity into uniswaps ModuleAccount
	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Sender, ModuleName, sdk.NewCoins(nativeCoin, coinDeposited))
	if err != nil {
		return err
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
	// check if exchange exists
	coinLiquidity, err := k.GetExchange(ctx, msg.ExchangeDenom)
	if err != nil {
		panic(fmt.Sprintf("error retrieving total liquidity for denomination: %s", msg.ExchangeDenom))
	}

	nativeLiquidity, err := k.GetExchange(ctx, NativeAsset)
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
	exchangeCoin = sdk.NewCoin(msg.ExchangeDenom, coinWithdrawn)

	// transfer withdrawn liquidity from uniswaps ModuleAccount to sender's account
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, msg.Sender, ModuleName, sdk.NewCoins(nativeCoin, coinDeposited))
	if err != nil {
		return err
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
