package service

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"

	"cosmossdk.io/tools/rosetta/lib/errors"
	crgtypes "cosmossdk.io/tools/rosetta/lib/types"
)

// Block gets the transactions in the given block
func (on OnlineNetwork) Block(ctx context.Context, request *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	var (
		blockResponse crgtypes.BlockTransactionsResponse
		err           error
	)

	// When fetching data by BlockIdentifier, it may be possible to only specify the index or hash.
	// If neither property is specified, it is assumed that the client is making a request at the current block.
	switch {
	case request.BlockIdentifier == nil: // unlike AccountBalance(), BlockIdentifer is mandatory by spec 1.4.10.
		err := errors.WrapError(errors.ErrBadArgument, "block identifier needs to be specified")
		return nil, errors.ToRosetta(err)

	case request.BlockIdentifier.Hash != nil:
		blockResponse, err = on.client.BlockTransactionsByHash(ctx, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}
	case request.BlockIdentifier.Index != nil:
		blockResponse, err = on.client.BlockTransactionsByHeight(ctx, request.BlockIdentifier.Index)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}

	default:
		// both empty
		blockResponse, err = on.client.BlockTransactionsByHeight(ctx, nil)
		if err != nil {
			return nil, errors.ToRosetta(err)
		}
	}

	// Both of index and hash can be specified in reuqest, so make sure they are not mismatching.
	if request.BlockIdentifier.Index != nil && *request.BlockIdentifier.Index != blockResponse.Block.Index {
		err := errors.WrapError(errors.ErrBadArgument, "mismatching index")
		return nil, errors.ToRosetta(err)
	}

	if request.BlockIdentifier.Hash != nil && *request.BlockIdentifier.Hash != blockResponse.Block.Hash {
		err := errors.WrapError(errors.ErrBadArgument, "mismatching hash")
		return nil, errors.ToRosetta(err)
	}

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       blockResponse.Block,
			ParentBlockIdentifier: blockResponse.ParentBlock,
			Timestamp:             blockResponse.MillisecondTimestamp,
			Transactions:          blockResponse.Transactions,
			Metadata:              nil,
		},
		OtherTransactions: nil,
	}, nil
}

// BlockTransaction gets the given transaction in the specified block, we do not need to check the block itself too
// due to the fact that CometBFT achieves instant finality
func (on OnlineNetwork) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	tx, err := on.client.GetTx(ctx, request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &types.BlockTransactionResponse{
		Transaction: tx,
	}, nil
}
