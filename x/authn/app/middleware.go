package app

import (
	"context"

	"github.com/cosmos/cosmos-sdk/core/app_config"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/core/module/app"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/authn"
)

type validateMemoMiddlewareHandler struct {
	*authn.ValidateMemoMiddleware
}

func (v validateMemoMiddlewareHandler) validate(tx tx.Tx) error {
	memoLength := len(tx.Body.Memo)
	if uint64(memoLength) > v.MaxMemoCharacters {
		return sdkerrors.Wrapf(sdkerrors.ErrMemoTooLarge,
			"maximum number of characters is %d but received %d characters",
			v.MaxMemoCharacters, memoLength,
		)
	}

	return nil
}

func (v validateMemoMiddlewareHandler) OnCheckTx(ctx context.Context, tx tx.Tx, req abci.RequestCheckTx, next app_config.TxHandler) (abci.ResponseCheckTx, error) {
	err := v.validate(tx)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return next.CheckTx(ctx, tx, req)
}

func (v validateMemoMiddlewareHandler) OnDeliverTx(ctx context.Context, tx tx.Tx, req abci.RequestDeliverTx, next app_config.TxHandler) (abci.ResponseDeliverTx, error) {
	err := v.validate(tx)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return next.DeliverTx(ctx, tx, req)
}

var _ app.TxMiddleware = validateMemoMiddlewareHandler{}
