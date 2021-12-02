package middleware

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ tx.Handler = mempoolFeeTxHandler{}

type mempoolFeeTxHandler struct {
	next tx.Handler
}

// MempoolFeeMiddleware will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, middleware returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next middleware
// CONTRACT: Tx must implement FeeTx to use MempoolFeeMiddleware
func MempoolFeeMiddleware(txh tx.Handler) tx.Handler {
	return mempoolFeeTxHandler{
		next: txh,
	}
}

// CheckTx implements tx.Handler.CheckTx. It is responsible for determining if a
// transaction's fees meet the required minimum of the processing node. Note, a
// node can have zero fees set as the minimum. If non-zero minimum fees are set
// and the transaction does not meet the minimum, the transaction is rejected.
//
// Recall, a transaction's fee is determined by ceil(minGasPrice * gasLimit).
func (txh mempoolFeeTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	feeTx, ok := req.Tx.(sdk.FeeTx)
	if !ok {
		return tx.Response{}, tx.ResponseCheckTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	minGasPrices := sdkCtx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return tx.Response{}, tx.ResponseCheckTx{}, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}
	}

	return txh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh mempoolFeeTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return txh.next.DeliverTx(ctx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (txh mempoolFeeTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return txh.next.SimulateTx(ctx, req)
}

var _ tx.Handler = deductFeeTxHandler{}

type deductFeeTxHandler struct {
	accountKeeper  AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper FeegrantKeeper
	next           tx.Handler
}

// DeductFeeMiddleware deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next middleware if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use deductFeeTxHandler
func DeductFeeMiddleware(ak AccountKeeper, bk types.BankKeeper, fk FeegrantKeeper) tx.Middleware {
	return func(txh tx.Handler) tx.Handler {
		return deductFeeTxHandler{
			accountKeeper:  ak,
			bankKeeper:     bk,
			feegrantKeeper: fk,
			next:           txh,
		}
	}
}

func (dfd deductFeeTxHandler) checkDeductFee(ctx context.Context, sdkTx sdk.Tx) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(sdkCtx, feeGranter, feePayer, fee, sdkTx.GetMsgs())

			if err != nil {
				return sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(sdkCtx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err := DeductFees(dfd.bankKeeper, sdkCtx, deductFeesFromAcc, feeTx.GetFee())
		if err != nil {
			return err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	sdkCtx.EventManager().EmitEvents(events)

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (dfd deductFeeTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	if err := dfd.checkDeductFee(ctx, req.Tx); err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return dfd.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (dfd deductFeeTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := dfd.checkDeductFee(ctx, req.Tx); err != nil {
		return tx.Response{}, err
	}

	return dfd.next.DeliverTx(ctx, req)
}

func (dfd deductFeeTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := dfd.checkDeductFee(ctx, req.Tx); err != nil {
		return tx.Response{}, err
	}

	return dfd.next.SimulateTx(ctx, req)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
