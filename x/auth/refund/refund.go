package refund

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type HandlerOptions struct {
	AccountKeeper  keeper.AccountKeeper
	BankKeeper     types.BankKeeper
	FeegrantKeeper ante.FeegrantKeeper
}

type GasRefundDecorator struct {
	ak keeper.AccountKeeper
	bk types.BankKeeper
	fk ante.FeegrantKeeper
}

func (grd GasRefundDecorator) GasRefundHandler(ctx sdk.Context, tx sdk.Tx, sim bool) (err error) {

	currentGasMeter := ctx.GasMeter()
	TempGasMeter := sdk.NewInfiniteGasMeter()
	ctx = ctx.WithGasMeter(TempGasMeter)

	defer func() {
		ctx = ctx.WithGasMeter(currentGasMeter)
	}()

	gasLimit := currentGasMeter.Limit()
	gasUsed := currentGasMeter.GasConsumed()

	if gasUsed >= gasLimit {
		return nil
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := grd.ak.GetModuleAddress(types.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
	}

	fees := feeTx.GetFee()
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()

	deductFeesFrom := feePayer

	if feeGranter != nil {
		if grd.fk == nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := grd.fk.UseGrantedFees(ctx, feeGranter, feePayer, fees)

			if err != nil {
				return sdkerrors.Wrapf(err, "%s not allowed to pay fees from %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := grd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	gas := feeTx.GetGas()
	gasFees := make(sdk.Coins, len(fees))

	for i, fee := range fees {
		gasPrice := new(big.Int).Div(fee.Amount.BigInt(), new(big.Int).SetUint64(gas))
		gasConsumed := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasUsed))
		gasCost := sdk.NewCoin(fee.Denom, sdk.NewIntFromBigInt(gasConsumed))
		gasRefund := fee.Sub(gasCost)
		gasFees[i] = gasRefund
	}

	err = RefundFees(grd.bk, ctx, deductFeesFromAcc.GetAddress(), gasFees)
	if err != nil {
		return err
	}

	return nil
}

func NewGasRefundDecorator(ak keeper.AccountKeeper, bk types.BankKeeper, fk ante.FeegrantKeeper) sdk.GasRefundHandler {

	grd := GasRefundDecorator{
		ak: ak,
		bk: bk,
		fk: fk,
	}

	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (err error) {
		return grd.GasRefundHandler(ctx, tx, simulate)
	}
}

func NewGasRefundHandler(options HandlerOptions) sdk.GasRefundHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (err error) {
		gasRefundHandler := NewGasRefundDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper)
		return gasRefundHandler(ctx, tx, sim)
	}
}

func RefundFees(bankKeeper types.BankKeeper, ctx sdk.Context, acc sdk.AccAddress, refundFees sdk.Coins) error {

	if !refundFees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid refund fee amount: %s", refundFees)
	}

	err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.FeeCollectorName, acc, refundFees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
