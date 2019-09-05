package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	err "github.com/cosmos/cosmos-sdk/types/errors"
	errs "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// ValidateBasicDecorator will call tx.ValidateBasic and return any non-nil error
// If ValidateBasic passes, decorator calls next AnteHandler in chain
type ValidateBasicDecorator struct{}

func NewValidateBasicDecorator() ValidateBasicDecorator {
	return ValidateBasicDecorator{}
}

func (vbd ValidateBasicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := tx.ValidateBasic(); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ValidateMemoDecorator will validate memo given the parameters passed in
// If memo is too large decorator returns with error, otherwise call next AnteHandler
type ValidateMemoDecorator struct {
	ak keeper.AccountKeeper
}

func NewValidateMemoDecorator(ak keeper.AccountKeeper) ValidateMemoDecorator {
	return ValidateMemoDecorator{
		ak: ak,
	}
}

func (vmd ValidateMemoDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		return ctx, errs.Wrap(errs.ErrTxDecode, "Tx must be a StdTx")
	}

	params := vmd.ak.GetParams(ctx)

	memoLength := len(stdTx.GetMemo())
	if uint64(memoLength) > params.MaxMemoCharacters {
		return ctx, err.Wrapf(err.ErrMemoTooLarge,
			"maximum number of characters is %d but received %d characters",
			params.MaxMemoCharacters, memoLength,
		)
	}

	return next(ctx, tx, simulate)
}

// ConsumeGasForTxSizeDecorator will take in parameters and consume gas proportional to the size of tx
// before calling next AnteHandler
type ConsumeGasForTxSizeDecorator struct {
	ak keeper.AccountKeeper
}

func NewConsumeGasForTxSizeDecorator(ak keeper.AccountKeeper) ConsumeGasForTxSizeDecorator {
	return ConsumeGasForTxSizeDecorator{
		ak: ak,
	}
}

func (cgts ConsumeGasForTxSizeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	params := cgts.ak.GetParams(ctx)
	ctx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(ctx.TxBytes())), "txSize")

	return next(ctx, tx, simulate)
}
