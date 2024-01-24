package txvalidator

import (
	"bytes"
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxFeeChecker check if the provided fee is enough and returns the effective fee and tx priority,
// the effective fee should be deducted later, and the priority should be returned in abci response.
type TxFeeChecker func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, error)

// DeductFeeDecorator deducts fees from the fee payer. The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	accountKeeper  feegrant.AccountKeeper
	bankKeeper     feegrant.BankKeeper
	feegrantKeeper keeper.Keeper
	txFeeChecker   TxFeeChecker
}

func NewDeductFeeDecorator(ak feegrant.AccountKeeper, bk feegrant.BankKeeper, fk keeper.Keeper, tfc TxFeeChecker) DeductFeeDecorator {
	if tfc == nil {
		tfc = checkTxFeeWithValidatorMinGasPrices
	}

	return DeductFeeDecorator{
		accountKeeper:  ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		txFeeChecker:   tfc,
	}
}

func (dfd DeductFeeDecorator) TxValidator(ctx context.Context, tx transaction.Tx) error {
	sdkTx := sdk.ServerTxToSDKTx(tx)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if sdkCtx.ExecMode() != sdk.ExecModeSimulate && sdkCtx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var err error
	fee := feeTx.GetFee()
	if sdkCtx.ExecMode() != sdk.ExecModeSimulate {
		fee, err = dfd.txFeeChecker(sdkCtx, sdkTx)
		if err != nil {
			return err
		}
	}

	if err := dfd.checkDeductFee(sdkCtx, sdkTx, fee); err != nil {
		return err
	}

	return err
}

func (dfd DeductFeeDecorator) checkDeductFee(ctx sdk.Context, sdkTx sdk.Tx, fee sdk.Coins) error {
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		feeGranterAddr := sdk.AccAddress(feeGranter)
		if !bytes.Equal(feeGranterAddr, feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranterAddr, feePayer, fee, sdkTx.GetMsgs())
			if err != nil {
				return errorsmod.Wrapf(err, "%s does not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranterAddr
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

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, sdk.AccAddress(deductFeesFrom).String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper feegrant.BankKeeper, ctx context.Context, acc sdk.AccountI, fees sdk.Coins) error {
	if !fees.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
