package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// DeductGrantedFeeDecorator deducts fees from fee_payer or fee_granter (if exists a valid fee allowance) of the tx
// If the fee_payer or fee_granter does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement GrantedFeeTx interface to use DeductGrantedFeeDecorator
type DeductGrantedFeeDecorator struct {
	ak types.AccountKeeper
	k  *feegrantkeeper.Keeper
	bk types.BankKeeper
}

func NewDeductGrantedFeeDecorator(ak types.AccountKeeper, bk types.BankKeeper, k *feegrantkeeper.Keeper) DeductGrantedFeeDecorator {
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

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	// ensure the grant is allowed, if we request a different fee payer
	if feeGranter != nil && !feeGranter.Equals(feePayer) {
		err := d.k.UseGrantedFees(ctx, feeGranter, feePayer, fee)
		if err != nil {
			return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
		}

	}

	return next(ctx, tx, simulate)
}
