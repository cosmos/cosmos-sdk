package ante

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FeeRecipientModule holds the module name most recently passed to
// WithFeeRecipientModule (default: fee_collector). It is what DeductFees sends
// fees to. It does NOT say where every DeductFeeDecorator sends fees: each
// decorator deducts to the recipient it was configured with, so an app that
// builds several decorators can have several recipients. Modules that need to
// know where an app's fees go should use a FeeRecipients registry.
var FeeRecipientModule = types.FeeCollectorName

// FeeRecipients records the fee recipient of every DeductFeeDecorator an app
// installs. It is created per app (not per process), and only decorators passed
// to Install are recorded, so decorators that are built but never installed,
// or that belong to another app in the same process, are not counted.
type FeeRecipients struct {
	mu      sync.Mutex
	modules map[string]struct{}
}

// NewFeeRecipients returns an empty FeeRecipients registry.
func NewFeeRecipients() *FeeRecipients {
	return &FeeRecipients{modules: map[string]struct{}{}}
}

// Install records the recipient of the given decorator value and returns the
// decorator unchanged, so it can wrap the decorator where it is added to the
// ante chain. Install on a nil registry is a no-op.
func (r *FeeRecipients) Install(dfd DeductFeeDecorator) DeductFeeDecorator {
	if r == nil {
		return dfd
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules[dfd.feeRecipientModule] = struct{}{}
	return dfd
}

// Modules returns the sorted module names that the installed decorators
// deduct fees to.
func (r *FeeRecipients) Modules() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	modules := make([]string, 0, len(r.modules))
	for m := range r.modules {
		modules = append(modules, m)
	}
	sort.Strings(modules)
	return modules
}

// TxFeeChecker check if the provided fee is enough and returns the effective fee and tx priority,
// the effective fee should be deducted later, and the priority should be returned in abci response.
type TxFeeChecker func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error)

// DeductFeeDecorator deducts fees from the fee payer. The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	accountKeeper      AccountKeeper
	bankKeeper         types.BankKeeper
	feegrantKeeper     FeegrantKeeper
	txFeeChecker       TxFeeChecker
	feeRecipientModule string
}

func NewDeductFeeDecorator(ak AccountKeeper, bk types.BankKeeper, fk FeegrantKeeper, tfc TxFeeChecker) DeductFeeDecorator {
	if tfc == nil {
		tfc = checkTxFeeWithValidatorMinGasPrices
	}

	return DeductFeeDecorator{
		accountKeeper:      ak,
		bankKeeper:         bk,
		feegrantKeeper:     fk,
		txFeeChecker:       tfc,
		feeRecipientModule: types.FeeCollectorName,
	}
}

// WithFeeRecipientModule sets the module account that receives deducted fees.
// By default this is the fee_collector. Use this to redirect fees to a
// different module (e.g. a distribution module that manages its own balance).
func (dfd DeductFeeDecorator) WithFeeRecipientModule(moduleName string) DeductFeeDecorator {
	if moduleName == "" {
		panic("fee recipient module name cannot be empty")
	}
	dfd.feeRecipientModule = moduleName
	FeeRecipientModule = moduleName
	return dfd
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var (
		priority int64
		err      error
	)

	fee := feeTx.GetFee()
	if !simulate {
		fee, priority, err = dfd.txFeeChecker(ctx, tx)
		if err != nil {
			return ctx, err
		}
	}
	if err := dfd.checkDeductFee(ctx, tx, fee); err != nil {
		return ctx, err
	}

	newCtx := ctx.WithPriority(priority)

	return next(newCtx, tx, simulate)
}

func (dfd DeductFeeDecorator) checkDeductFee(ctx sdk.Context, sdkTx sdk.Tx, fee sdk.Coins) error {
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(dfd.feeRecipientModule); addr == nil {
		return fmt.Errorf("fee recipient module account (%s) has not been set", dfd.feeRecipientModule)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		feeGranterAddr := sdk.AccAddress(feeGranter)

		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !bytes.Equal(feeGranterAddr, feePayer) {
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
		err := deductFeesToModule(dfd.bankKeeper, ctx, deductFeesFromAcc, fee, dfd.feeRecipientModule)
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

// DeductFees deducts fees from the given account and sends them to the
// module configured via FeeRecipientModule.
func DeductFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc sdk.AccountI, fees sdk.Coins) error {
	return deductFeesToModule(bankKeeper, ctx, acc, fees, FeeRecipientModule)
}

// deductFeesToModule deducts fees from the given account and sends them to
// the given module account.
func deductFeesToModule(bankKeeper types.BankKeeper, ctx sdk.Context, acc sdk.AccountI, fees sdk.Coins, recipientModule string) error {
	if !fees.IsValid() {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), recipientModule, fees)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "%s", err.Error())
	}

	return nil
}
