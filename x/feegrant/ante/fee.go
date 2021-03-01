package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// DeductGrantedFeeDecorator deducts fees from fee_payer or fee_granter (if exists a valid fee allowance) of the tx
// If the fee_payer or fee_granter does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement GrantedFeeTx interface to use DeductGrantedFeeDecorator
type DeductGrantedFeeDecorator struct {
	ak types.AccountKeeper
	k  keeper.Keeper
	bk types.BankKeeper
}

func NewDeductGrantedFeeDecorator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) DeductGrantedFeeDecorator {
	return DeductGrantedFeeDecorator{
		ak: ak,
		k:  k,
		bk: bk,
	}
}

// AnteHandle performs a decorated ante-handler responsible for deducting transaction
// fees. Fees will be deducted from the account designated by the FeePayer on a
// transaction by default. However, if the fee payer differs from the transaction
// signer, the handler will check if a fee grant has been authorized. If the
// transaction's signer does not exist, it will be created.
func (d DeductGrantedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a GrantedFeeTx")
	}

	// sanity check from DeductFeeDecorator
	if addr := d.ak.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", authtypes.FeeCollectorName))
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	deductFeesFrom := feePayer

	// ensure the grant is allowed, if we request a different fee payer
	if feeGranter != nil && !feeGranter.Equals(feePayer) {
		err := d.k.UseGrantedFees(ctx, feeGranter, feePayer, fee, tx.GetMsgs())
		if err != nil {
			return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
		}

		deductFeesFrom = feeGranter
	}

	// now, either way, we know that we are authorized to deduct the fees from the deductFeesFrom account
	deductFeesFromAcc := d.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	// move on if there is no fee to deduct
	if fee.IsZero() {
		return next(ctx, tx, simulate)
	}

	// deduct fee if non-zero
	err = authante.DeductFees(d.bk, ctx, deductFeesFromAcc, fee)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
