package handlers

import (
	"context"
	"errors"
	"fmt"

	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

type DefaultProposalHandler[T transaction.Tx] struct {
	mempool    mempool.Mempool[T]
	txSelector TxSelector[T]
}

func NewDefaultProposalHandler[T transaction.Tx](mp mempool.Mempool[T]) *DefaultProposalHandler[T] {
	return &DefaultProposalHandler[T]{
		mempool:    mp,
		txSelector: NewDefaultTxSelector[T](),
	}
}

func (h *DefaultProposalHandler[T]) PrepareHandler() appmanager.PrepareHandler[T] {
	return func(ctx context.Context, app appmanager.AppManager[T], rm store.ReaderMap, txs []T, req proto.Message) ([]T, error) {

		abciReq, ok := req.(*abci.RequestPrepareProposal)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", req)
		}

		var maxBlockGas uint64

		res, err := app.QueryWithState(ctx, rm, &consensusv1.QueryParamsRequest{})
		if err != nil {
			return nil, err
		}

		paramsResp, ok := res.(*consensusv1.QueryParamsResponse)
		if !ok {
			return nil, fmt.Errorf("failed to query consensus params")
		}

		if b := paramsResp.GetParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		defer h.txSelector.Clear()

		// TODO: can we assume nil mempool is NoOp?
		// If the mempool is nil or NoOp we simply return the transactions
		// requested from CometBFT, which, by default, should be in FIFO order.
		//
		// Note, we still need to ensure the transactions returned respect req.MaxTxBytes.
		if h.mempool == nil {
			for _, tx := range txs {
				stop := h.txSelector.SelectTxForProposal(ctx, uint64(abciReq.MaxTxBytes), maxBlockGas, tx)
				if stop {
					break
				}
			}

			return h.txSelector.SelectedTxs(ctx), nil
		}

		iterator := h.mempool.Select(ctx, txs)
		for iterator != nil {
			memTx := iterator.Tx()

			// NOTE: Since transaction verification was already executed in CheckTx,
			// which calls mempool.Insert, in theory everything in the pool should be
			// valid. But some mempool implementations may insert invalid txs, so we
			// check again.
			_, err := app.ValidateTx(ctx, memTx, corecontext.ExecModePrepareProposal)

			if err != nil {
				err := h.mempool.Remove(memTx)
				if err != nil && !errors.Is(err, mempool.ErrTxNotFound) {
					return nil, err
				}
			} else {
				stop := h.txSelector.SelectTxForProposal(ctx, uint64(abciReq.MaxTxBytes), maxBlockGas, memTx)
				if stop {
					break
				}
			}

			iterator = iterator.Next()
		}

		return h.txSelector.SelectedTxs(ctx), nil

	}
}

func (h *DefaultProposalHandler[T]) ProcessHandler() appmanager.ProcessHandler[T] {
	return func(ctx context.Context, app appmanager.AppManager[T], txs []T, rm store.ReaderMap, req proto.Message) error {
		// If the mempool is nil we simply return ACCEPT,
		// because PrepareProposal may have included txs that could fail verification.
		if h.mempool == nil {
			return nil
		}

		// TODO: not using this request for now
		_, ok := req.(*abci.RequestProcessProposal)
		if !ok {
			return fmt.Errorf("invalid request type: %T", req)
		}

		res, err := app.QueryWithState(ctx, rm, &consensusv1.QueryParamsRequest{})
		if err != nil {
			return err
		}

		paramsResp, ok := res.(*consensusv1.QueryParamsResponse)
		if !ok {
			return fmt.Errorf("failed to query consensus params")
		}

		var maxBlockGas uint64
		if b := paramsResp.GetParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		var totalTxGas uint64
		for _, tx := range txs {
			_, err := app.ValidateTx(ctx, tx, corecontext.ExecModePrepareProposal)
			if err != nil {
				return fmt.Errorf("failed to validate tx: %w", err)
			}

			if maxBlockGas > 0 {
				totalTxGas += tx.GetGasLimit()
				if totalTxGas > uint64(maxBlockGas) {
					return fmt.Errorf("total tx gas %d exceeds max block gas %d", totalTxGas, maxBlockGas)
				}
			}
		}

		return nil
	}
}
