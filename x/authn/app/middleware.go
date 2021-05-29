package app

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/core/tx/module"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/authn"
)

type validateMemoMiddlewareHandler struct {
	module.TxHandler
	*authn.ValidateMemoMiddleware
}

func (v validateMemoMiddlewareHandler) CheckTx(ctx context.Context, tx tx.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	err := v.validate(tx)
	if err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return v.TxHandler.CheckTx(ctx, tx, req)
}

func (v validateMemoMiddlewareHandler) DeliverTx(ctx context.Context, tx tx.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	err := v.validate(tx)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return v.TxHandler.DeliverTx(ctx, tx, req)
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
