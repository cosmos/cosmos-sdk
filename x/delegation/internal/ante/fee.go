package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	// we depend on the auth module internals... maybe some more of this can be exported?
	// but things like `x/auth/types/FeeCollectorName` are quite clearly tied to it
	authAnte "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/keeper"

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
	base authAnte.DeductFeeDecorator
	ak   authKeeper.AccountKeeper
	dk   keeper.Keeper
	sk   authTypes.SupplyKeeper
}

func NewDeductDelegatedFeeDecorator(ak authKeeper.AccountKeeper, sk authTypes.SupplyKeeper, dk keeper.Keeper) DeductDelegatedFeeDecorator {
	return DeductDelegatedFeeDecorator{
		base: authAnte.NewDeductFeeDecorator(ak, sk),
		ak:   ak,
		dk:   dk,
		sk:   sk,
	}
}

func (d DeductDelegatedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// make sure there is a delegation, if not, default to the standard DeductFeeDecorator behavior
	var granter sdk.AccAddress
	delTx, ok := tx.(DelegatedFeeTx)
	if ok {
		granter = delTx.GetFeeAccount()
	}
	if granter == nil {
		// just defer to the basic DeductFeeHandler
		return d.base.AnteHandle(ctx, tx, simulate, next)
	}

	// short-circuit on zero fee
	fee := delTx.GetFee()
	if fee.IsZero() {
		return next(ctx, tx, simulate)
	}

	// ensure the delegation is allowed
	grantee := delTx.FeePayer()
	allowed := d.dk.UseDelegatedFees(ctx, granter, grantee, fee)
	if !allowed {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s not allowed to pay fees from %s", grantee, granter)
	}

	// now deduct fees from the granter
	feePayerAcc := d.ak.GetAccount(ctx, granter)
	if feePayerAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "granter address: %s does not exist", granter)
	}
	err = authAnte.DeductFees(d.sk, ctx, feePayerAcc, fee)
	if err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}
