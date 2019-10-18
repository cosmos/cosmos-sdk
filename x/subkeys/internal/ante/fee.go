package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// we depend on the auth module internals... maybe some more of this can be exported?
	// but things like `x/auth/types/FeeCollectorName` are quite clearly tied to it
	authAnte "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/subkeys/internal/keeper"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ DelegatedFeeTx = (*authTypes.StdTx)(nil) // assert StdTx implements DelegatedFeeTx
)

// DelegatedFeeTx defines the interface to be implemented by Tx to use the DelegatedFeeDecorator
type DelegatedFeeTx interface {
	authAnte.FeeTx
	GetFeeAccount() sdk.AccAddress
}

// DeductDelegatedFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement DelegatedFeeTx interface to use DeductDelegatedFeeDecorator
type DeductDelegatedFeeDecorator struct {
	ak authKeeper.AccountKeeper
	dk keeper.Keeper
	sk authTypes.SupplyKeeper
}

func NewDeductDelegatedFeeDecorator(ak authKeeper.AccountKeeper, sk authTypes.SupplyKeeper, dk keeper.Keeper) DeductDelegatedFeeDecorator {
	return DeductDelegatedFeeDecorator{
		ak: ak,
		dk: dk,
		sk: sk,
	}
}

func (d DeductDelegatedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(authAnte.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// sanity check from DeductFeeDecorator
	if addr := d.sk.GetModuleAddress(authTypes.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", authTypes.FeeCollectorName))
	}

	// see if there is a delegation
	fee := feeTx.GetFee()
	var feePayer sdk.AccAddress
	if delTx, ok := tx.(DelegatedFeeTx); ok {
		feePayer = delTx.GetFeeAccount()
	}

	txSigner := feeTx.FeePayer()
	if feePayer == nil {
		// if this is not explicitly set, use the first signer as always
		feePayer = txSigner
	} else {
		// ensure the delegation is allowed
		allowed := d.dk.UseDelegatedFees(ctx, feePayer, txSigner, fee)
		if !allowed {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s not allowed to pay fees from %s", txSigner, feePayer)
		}
	}

	// now, either way, we know that we are authorized to deduct the fees from the feePayer account
	feePayerAcc := d.ak.GetAccount(ctx, feePayer)
	if feePayerAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", feePayer)
	}

	// deduct fee if non-zero
	if fee.IsZero() {
		return next(ctx, tx, simulate)
	}
	err = authAnte.DeductFees(d.sk, ctx, feePayerAcc, fee)
	if err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}
