package middleware

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TxFeeChecker check if the provided fee is enough and returns the effective fee and tx priority,
// the effective fee should be deducted later, and the priority should be returned in abci response.
type TxFeeChecker func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error)

var _ tx.Handler = deductFeeTxHandler{}

type deductFeeTxHandler struct {
	accountKeeper  AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper FeegrantKeeper
	txFeeChecker   TxFeeChecker
	next           tx.Handler
}

// DeductFeeMiddleware deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next middleware if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use deductFeeTxHandler
func DeductFeeMiddleware(ak AccountKeeper, bk types.BankKeeper, fk FeegrantKeeper, tfc TxFeeChecker) tx.Middleware {
	if tfc == nil {
		tfc = checkTxFeeWithValidatorMinGasPrices
	}
	return func(txh tx.Handler) tx.Handler {
		return deductFeeTxHandler{
			accountKeeper:  ak,
			bankKeeper:     bk,
			feegrantKeeper: fk,
			txFeeChecker:   tfc,
			next:           txh,
		}
	}
}

func (dfd deductFeeTxHandler) checkDeductFee(ctx sdk.Context, sdkTx sdk.Tx, fee sdk.Coins) error {
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return fmt.Errorf("Fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, sdkTx.GetMsgs())
			if err != nil {
				return sdkerrors.Wrapf(err, "%s does not not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !fee.IsZero() {
		err := DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, fee)
		if err != nil {
			return err
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (dfd deductFeeTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	fee, priority, err := dfd.txFeeChecker(sdkCtx, req.Tx)
	if err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}
	if err := dfd.checkDeductFee(sdkCtx, req.Tx, fee); err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	res, checkRes, err := dfd.next.CheckTx(ctx, req, checkReq)
	checkRes.Priority = priority

	return res, checkRes, err
}

// DeliverTx implements tx.Handler.DeliverTx.
func (dfd deductFeeTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	fee, _, err := dfd.txFeeChecker(sdkCtx, req.Tx)
	if err != nil {
		return tx.Response{}, err
	}
	if err := dfd.checkDeductFee(sdkCtx, req.Tx, fee); err != nil {
		return tx.Response{}, err
	}

	return dfd.next.DeliverTx(ctx, req)
}

func (dfd deductFeeTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	fee, _, err := dfd.txFeeChecker(sdkCtx, req.Tx)
	if err != nil {
		return tx.Response{}, err
	}
	if err := dfd.checkDeductFee(sdkCtx, req.Tx, fee); err != nil {
		return tx.Response{}, err
	}

	return dfd.next.SimulateTx(ctx, req)
}

// Deprecated: DeductFees deducts fees from the given account.
// This function will be private in the next release.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc types.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return sdkerrors.ErrInsufficientFunds.Wrap(err.Error())
	}

	return nil
}
