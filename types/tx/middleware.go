package tx

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxHandler interface {
	CheckTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	DeliverTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
}

type TxMiddleware func(TxHandler) TxHandler
