package handlers

import (
	"context"
	"errors"
	"fmt"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/cometbft/mempool"
	consensustypes "cosmossdk.io/x/consensus/types"
)

type AppManager[T transaction.Tx] interface {
	ValidateTx(ctx context.Context, tx T) (server.TxResult, error)
	Query(ctx context.Context, version uint64, request transaction.Msg) (response transaction.Msg, err error)
}

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

func (h *DefaultProposalHandler[T]) PrepareHandler() PrepareHandler[T] {
	return func(ctx context.Context, app AppManager[T], codec transaction.Codec[T], req *abci.PrepareProposalRequest) ([]T, error) {
		var maxBlockGas uint64

		res, err := app.Query(ctx, 0, &consensustypes.QueryParamsRequest{})
		if err != nil {
			return nil, err
		}

		paramsResp, ok := res.(*consensustypes.QueryParamsResponse)
		if !ok {
			return nil, fmt.Errorf("unexpected consensus params response type; expected: %T, got: %T", &consensustypes.QueryParamsResponse{}, res)
		}

		if b := paramsResp.GetParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		txs := decodeTxs(codec, req.Txs)

		defer h.txSelector.Clear()

		// If the mempool is nil or NoOp we simply return the transactions
		// requested from CometBFT, which, by default, should be in FIFO order.
		//
		// Note, we still need to ensure the transactions returned respect req.MaxTxBytes.
		_, isNoOp := h.mempool.(mempool.NoOpMempool[T])
		if h.mempool == nil || isNoOp {
			for _, tx := range txs {
				stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx)
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
			_, err := app.ValidateTx(ctx, memTx)
			if err != nil {
				err := h.mempool.Remove(memTx)
				if err != nil && !errors.Is(err, mempool.ErrTxNotFound) {
					return nil, err
				}
			} else {
				stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, memTx)
				if stop {
					break
				}
			}

			iterator = iterator.Next()
		}

		return h.txSelector.SelectedTxs(ctx), nil
	}
}

func (h *DefaultProposalHandler[T]) ProcessHandler() ProcessHandler[T] {
	return func(ctx context.Context, app AppManager[T], codec transaction.Codec[T], req *abci.ProcessProposalRequest) error {
		// If the mempool is nil we simply return ACCEPT,
		// because PrepareProposal may have included txs that could fail verification.
		_, isNoOp := h.mempool.(mempool.NoOpMempool[T])
		if h.mempool == nil || isNoOp {
			return nil
		}

		res, err := app.Query(ctx, 0, &consensustypes.QueryParamsRequest{})
		if err != nil {
			return err
		}

		paramsResp, ok := res.(*consensustypes.QueryParamsResponse)
		if !ok {
			return fmt.Errorf("unexpected consensus params response type; expected: %T, got: %T", &consensustypes.QueryParamsResponse{}, res)
		}

		var maxBlockGas uint64
		if b := paramsResp.GetParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		// Decode request txs bytes
		// If there an tx decoded fail, return err
		var txs []T
		for _, tx := range req.Txs {
			decTx, err := codec.Decode(tx)
			if err != nil {
				return fmt.Errorf("failed to decode tx: %w", err)
			}
			txs = append(txs, decTx)
		}

		var totalTxGas uint64
		for _, tx := range txs {
			_, err := app.ValidateTx(ctx, tx)
			if err != nil {
				return fmt.Errorf("failed to validate tx: %w", err)
			}

			if maxBlockGas > 0 {
				gaslimit, err := tx.GetGasLimit()
				if err != nil {
					return errors.New("failed to get gas limit")
				}
				totalTxGas += gaslimit
				if totalTxGas > maxBlockGas {
					return fmt.Errorf("total tx gas %d exceeds max block gas %d", totalTxGas, maxBlockGas)
				}
			}
		}

		return nil
	}
}

// decodeTxs decodes the txs bytes into a decoded txs
// If there a fail decoding tx, remove from the list
// Used for prepare proposal
func decodeTxs[T transaction.Tx](codec transaction.Codec[T], txsBz [][]byte) []T {
	var txs []T
	for _, tx := range txsBz {
		decTx, err := codec.Decode(tx)
		if err != nil {
			continue
		}

		txs = append(txs, decTx)
	}
	return txs
}

// NoOpPrepareProposal defines a no-op PrepareProposal handler. It will always
// return the transactions sent by the client's request.
func NoOpPrepareProposal[T transaction.Tx]() PrepareHandler[T] {
	return func(ctx context.Context, app AppManager[T], codec transaction.Codec[T], req *abci.PrepareProposalRequest) ([]T, error) {
		return decodeTxs(codec, req.Txs), nil
	}
}

// NoOpProcessProposal defines a no-op ProcessProposal Handler. It will always
// return ACCEPT.
func NoOpProcessProposal[T transaction.Tx]() ProcessHandler[T] {
	return func(context.Context, AppManager[T], transaction.Codec[T], *abci.ProcessProposalRequest) error {
		return nil
	}
}

// NoOpExtendVote defines a no-op ExtendVote handler. It will always return an
// empty byte slice as the vote extension.
func NoOpExtendVote() ExtendVoteHandler {
	return func(context.Context, store.ReaderMap, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
		return &abci.ExtendVoteResponse{VoteExtension: []byte{}}, nil
	}
}

// NoOpVerifyVoteExtensionHandler defines a no-op VerifyVoteExtension handler. It
// will always return an ACCEPT status with no error.
func NoOpVerifyVoteExtensionHandler() VerifyVoteExtensionhandler {
	return func(context.Context, store.ReaderMap, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT}, nil
	}
}
