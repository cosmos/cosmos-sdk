package service

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"

	"cosmossdk.io/tools/rosetta/lib/errors"
)

// Mempool fetches the transactions contained in the mempool
func (on OnlineNetwork) Mempool(ctx context.Context, _ *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	txs, err := on.client.Mempool(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &types.MempoolResponse{
		TransactionIdentifiers: txs,
	}, nil
}

// MempoolTransaction fetches a single transaction in the mempool
// NOTE: it is not implemented yet
func (on OnlineNetwork) MempoolTransaction(ctx context.Context, request *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	tx, err := on.client.GetUnconfirmedTx(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &types.MempoolTransactionResponse{
		Transaction: tx,
	}, nil
}
