package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// we depend on the auth module internals... maybe some more of this can be exported?
	// but things like `x/auth/types/FeeCollectorName` are quite clearly tied to it
	authAnte "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/types/tx"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ GrantedFeeTx = (*tx.FeeGrantTx)(nil) // assert StdTx implements GrantedFeeTx
)

// GrantedFeeTx defines the interface to be implemented by Tx to use the GrantedFeeDecorator
type GrantedFeeTx interface {
	sdk.Tx
	GetGas() uint64
	GetFee() sdk.Coins
	FeePayer() sdk.AccAddress
	MainSigner() sdk.AccAddress
}

// DeductGrantedFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement GrantedFeeTx interface to use DeductGrantedFeeDecorator
type DeductGrantedFeeDecorator struct {
	ak authKeeper.AccountKeeper
	dk keeper.Keeper
	sk authTypes.SupplyKeeper
}

func NewDeductGrantedFeeDecorator(ak authKeeper.AccountKeeper, sk authTypes.SupplyKeeper, dk keeper.Keeper) DeductGrantedFeeDecorator {
	return DeductGrantedFeeDecorator{
		ak: ak,
		dk: dk,
		sk: sk,
	}
}

func (d DeductGrantedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(GrantedFeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a GrantedFeeTx")
	}

	// sanity check from DeductFeeDecorator
	if addr := d.sk.GetModuleAddress(authTypes.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", authTypes.FeeCollectorName))
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	txSigner := feeTx.MainSigner()

	// ensure the grant is allowed, if we request a different fee payer
	if !txSigner.Equals(feePayer) {
		allowed := d.dk.UseGrantedFees(ctx, feePayer, txSigner, fee)
		if !allowed {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s not allowed to pay fees from %s", txSigner, feePayer)
		}
		// if there was a valid grant, ensure that the txSigner account exists (we create it if needed)
		signerAcc := d.ak.GetAccount(ctx, txSigner)
		if signerAcc == nil {
			signerAcc = d.ak.NewAccountWithAddress(ctx, txSigner)
			d.ak.SetAccount(ctx, signerAcc)
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
