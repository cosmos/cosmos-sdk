package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

var (
	_ GrantedFeeTx = (*types.FeeGrantTx)(nil) // assert FeeGrantTx implements GrantedFeeTx
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
	ak types.AccountKeeper
	k  keeper.Keeper
	sk types.SupplyKeeper
}

func NewDeductGrantedFeeDecorator(ak types.AccountKeeper, sk types.SupplyKeeper, k keeper.Keeper) DeductGrantedFeeDecorator {
	return DeductGrantedFeeDecorator{
		ak: ak,
		k:  k,
		sk: sk,
	}
}

// AnteHandle performs a decorated ante-handler responsible for deducting transaction
// fees. Fees will be deducted from the account designated by the FeePayer on a
// transaction by default. However, if the fee payer differs from the transaction
// signer, the handler will check if a fee grant has been authorized. If the
// transaction's signer does not exist, it will be created.
func (d DeductGrantedFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(GrantedFeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a GrantedFeeTx")
	}

	// sanity check from DeductFeeDecorator
	if addr := d.sk.GetModuleAddress(auth.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", auth.FeeCollectorName))
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	txSigner := feeTx.MainSigner()

	// ensure the grant is allowed, if we request a different fee payer
	if !txSigner.Equals(feePayer) {
		err := d.k.UseGrantedFees(ctx, feePayer, txSigner, fee)
		if err != nil {
			return ctx, sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", txSigner, feePayer)
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

	// move on if there is no fee to deduct
	if fee.IsZero() {
		return next(ctx, tx, simulate)
	}

	// deduct fee if non-zero
	err = auth.DeductFees(d.sk, ctx, feePayerAcc, fee)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
