package ante

import (
	"bytes"
	"context"
	"fmt"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxFeeChecker checks if the provided fee is enough and returns the effective fee and tx priority.
// The effective fee should be deducted later, and the priority should be returned in the ABCI response.
type TxFeeChecker func(ctx context.Context, tx sdk.Tx) (sdk.Coins, int64, error)

// FeeTxValidator defines custom type used to represent deduct fee decorator
// which will be passed from x/auth/tx to x/auth module.
type FeeTxValidator interface {
	appmodulev2.TxValidator[sdk.Tx]

	SetMinGasPrices(sdk.DecCoins)
	SetFeegrantKeeper(FeegrantKeeper) FeeTxValidator
}

// DeductFeeDecorator deducts fees from the fee payer. The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees are successfully deducted.
// CONTRACT: The Tx must implement the FeeTx interface to use DeductFeeDecorator.
type DeductFeeDecorator struct {
	accountKeeper  AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper FeegrantKeeper
	txFeeChecker   TxFeeChecker

	// pointer to a separate state struct
	state *deductFeeState
}

// deductFeeState holds the mutable state needed across method calls
type deductFeeState struct {
	minGasPrices   sdk.DecCoins
	feeTx          sdk.FeeTx
	txPriority     int64
	deductFeesFrom []byte
	txFee          sdk.Coins
	execMode       transaction.ExecMode
}

func NewDeductFeeDecorator(ak AccountKeeper, bk types.BankKeeper, fk FeegrantKeeper, tfc TxFeeChecker) DeductFeeDecorator {
	dfd := DeductFeeDecorator{
		accountKeeper:  ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		txFeeChecker:   tfc,
		state:          &deductFeeState{}, // Initialize the state
	}

	if tfc == nil {
		dfd.txFeeChecker = dfd.checkTxFeeWithValidatorMinGasPrices
	}

	return dfd
}

// SetMinGasPrices sets the minimum-gas-prices value in the state of DeductFeeDecorator
func (dfd DeductFeeDecorator) SetMinGasPrices(minGasPrices sdk.DecCoins) {
	dfd.state.minGasPrices = minGasPrices
}

// SetFeegrantKeeper sets the feegrant keeper in DeductFeeDecorator
func (dfd DeductFeeDecorator) SetFeegrantKeeper(feegrantKeeper FeegrantKeeper) FeeTxValidator {
	dfd.feegrantKeeper = feegrantKeeper
	return dfd
}

// AnteHandle implements an AnteHandler decorator for the DeductFeeDecorator
func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, _ bool, next sdk.AnteHandler) (sdk.Context, error) {
	dfd.state.minGasPrices = ctx.MinGasPrices()

	if err := dfd.ValidateTx(ctx, tx); err != nil {
		return ctx, err
	}

	// TODO: emit this event in v2 after executing ValidateTx method
	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, dfd.state.txFee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, sdk.AccAddress(dfd.state.deductFeesFrom).String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	newCtx := ctx.WithPriority(dfd.state.txPriority)
	return next(newCtx, tx, false)
}

// ValidateTx implements an TxValidator for DeductFeeDecorator
func (dfd DeductFeeDecorator) ValidateTx(ctx context.Context, tx sdk.Tx) error {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must implement the FeeTx interface")
	}

	// update the state with the current transaction
	dfd.state.feeTx = feeTx

	dfd.state.execMode = dfd.accountKeeper.GetEnvironment().TransactionService.ExecMode(ctx)
	headerInfo := dfd.accountKeeper.GetEnvironment().HeaderService.HeaderInfo(ctx)

	if dfd.state.execMode != transaction.ExecModeSimulate && headerInfo.Height > 0 && dfd.state.feeTx.GetGas() == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	dfd.state.txFee = dfd.state.feeTx.GetFee()

	var err error

	if dfd.state.execMode != transaction.ExecModeSimulate {
		dfd.state.txFee, dfd.state.txPriority, err = dfd.txFeeChecker(ctx, tx)
		if err != nil {
			return err
		}
	}

	if err := dfd.checkDeductFee(ctx, tx, dfd.state.txFee); err != nil {
		return err
	}

	return nil
}

func (dfd DeductFeeDecorator) checkDeductFee(ctx context.Context, sdkTx sdk.Tx, fee sdk.Coins) error {
	addr := dfd.accountKeeper.GetModuleAddress(types.FeeCollectorName)
	if len(addr) == 0 {
		return fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	feePayer := dfd.state.feeTx.FeePayer()
	feeGranter := dfd.state.feeTx.FeeGranter()
	dfd.state.deductFeesFrom = feePayer

	// if feegranter set, deduct fee from feegranter account.
	// this works only when feegrant is enabled.
	if feeGranter != nil {
		feeGranterAddr := sdk.AccAddress(feeGranter)

		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !bytes.Equal(feeGranterAddr, feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranterAddr, feePayer, fee, sdkTx.GetMsgs())
			if err != nil {
				granterAddr, acErr := dfd.accountKeeper.AddressCodec().BytesToString(feeGranter)
				if acErr != nil {
					return errorsmod.Wrapf(err, "%s, feeGranter does not allow to pay fees", acErr.Error())
				}
				payerAddr, acErr := dfd.accountKeeper.AddressCodec().BytesToString(feePayer)
				if acErr != nil {
					return errorsmod.Wrapf(err, "%s, feeGranter does not allow to pay fees", acErr.Error())
				}
				return errorsmod.Wrapf(err, "%s does not allow to pay fees for %s", granterAddr, payerAddr)
			}
		}

		dfd.state.deductFeesFrom = feeGranterAddr
	}

	// deduct the fees
	if !fee.IsZero() {
		err := DeductFees(dfd.bankKeeper, ctx, dfd.state.deductFeesFrom, fee)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx context.Context, acc []byte, fees sdk.Coins) error {
	if !fees.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.AccAddress(acc), types.FeeCollectorName, fees)
	if err != nil {
		return fmt.Errorf("failed to deduct fees: %w", err)
	}

	return nil
}
