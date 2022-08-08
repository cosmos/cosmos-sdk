package ante

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TxFeeChecker checks if the provided fee is enough and returns the effective fee and tx priority.
// The effective fee should be deducted later, and the priority should be returned in the ABCI response.
type TxFeeChecker func(ctx context.Context, tx transaction.Tx) (sdk.Coins, int64, error)

// DeductFeeDecorator deducts fees from the fee payer. The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees are successfully deducted.
// CONTRACT: The Tx must implement the FeeTx interface to use DeductFeeDecorator.
type DeductFeeDecorator struct {
	accountKeeper  AccountKeeper
	bankKeeper     types.BankKeeper
	feegrantKeeper FeegrantKeeper
	txFeeChecker   TxFeeChecker
	minGasPrices   sdk.DecCoins
}

func NewDeductFeeDecorator(ak AccountKeeper, bk types.BankKeeper, fk FeegrantKeeper, tfc TxFeeChecker) *DeductFeeDecorator {
	dfd := &DeductFeeDecorator{
		accountKeeper:  ak,
		bankKeeper:     bk,
		feegrantKeeper: fk,
		txFeeChecker:   tfc,
		minGasPrices:   sdk.NewDecCoins(),
	}

	if tfc == nil {
		dfd.txFeeChecker = dfd.checkTxFeeWithValidatorMinGasPrices
	}

	return dfd
}

// SetMinGasPrices sets the minimum-gas-prices value in the state of DeductFeeDecorator
func (dfd *DeductFeeDecorator) SetMinGasPrices(minGasPrices sdk.DecCoins) {
	dfd.minGasPrices = minGasPrices
}

// AnteHandle implements an AnteHandler decorator for the DeductFeeDecorator
func (dfd *DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, _ bool, next sdk.AnteHandler) (sdk.Context, error) {
	dfd.minGasPrices = ctx.MinGasPrices()
	txPriority, err := dfd.innerValidateTx(ctx, tx)
	if err != nil {
		return ctx, err
	}

	newCtx := ctx.WithPriority(txPriority)
	return next(newCtx, tx, false)
}

func (dfd *DeductFeeDecorator) innerValidateTx(ctx context.Context, tx transaction.Tx) (priority int64, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return 0, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must implement the FeeTx interface")
	}

	if addr := dfd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return ctx, fmt.Errorf("Fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	fee := feeTx.GetFee()
	if execMode != transaction.ExecModeSimulate {
		fee, priority, err = dfd.txFeeChecker(ctx, tx)
		if err != nil {
			return 0, err
		}
	}

	if err := dfd.checkDeductFee(ctx, feeTx, fee); err != nil {
		return 0, err
	}

	return priority, nil
}

// ValidateTx implements an TxValidator for DeductFeeDecorator
// Note: This method is applicable only for transactions that implement the sdk.FeeTx interface.
func (dfd *DeductFeeDecorator) ValidateTx(ctx context.Context, tx transaction.Tx) error {
	_, err := dfd.innerValidateTx(ctx, tx)
	return err
}

func (dfd *DeductFeeDecorator) checkDeductFee(ctx context.Context, feeTx sdk.FeeTx, fee sdk.Coins) error {
	addr := dfd.accountKeeper.GetModuleAddress(types.FeeCollectorName)
	if len(addr) == 0 {
		return fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set, deduct fee from feegranter account.
	// this works only when feegrant is enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
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
		deductFeesFrom = feeGranter
	}

	// deduct the fees
	if !fee.IsZero() {
		if err := DeductFees(dfd.bankKeeper, ctx, deductFeesFrom, fee); err != nil {
			return err
		}
	}

	events := sdk.Events{
		sdk.NewEvent(sdk.EventTypeTx, sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String())),
		sdk.NewEvent(sdk.EventTypeTx, sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String())),
	}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees from the given account.
func DeductFees(bankKeeper types.BankKeeper, ctx context.Context, acc []byte, fees sdk.Coins) error {
	if !fees.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	if err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.FeeCollectorName, fees); err != nil {
		return fmt.Errorf("failed to deduct fees: %w", err)
	}

	return nil
}
