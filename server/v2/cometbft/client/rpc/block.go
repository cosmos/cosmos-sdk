package rpc

import (
	"context"
	"encoding/hex"
	"fmt"

	cmttypes "github.com/cometbft/cometbft/api/cometbft/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetChainHeight returns the current blockchain height.
func GetChainHeight(ctx context.Context, rpcClient CometRPC) (int64, error) {
	status, err := rpcClient.Status(ctx)
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// QueryBlocks performs a search for blocks based on BeginBlock and EndBlock
// events via the CometBFT RPC. A custom query may be passed as described below:
//
// To tell which events you want, you need to provide a query. query is a
// string, which has a form: "condition AND condition ..." (no OR at the
// 	moment). condition has a form: "key operation operand". key is a string with
// 	a restricted set of possible symbols ( \t\n\r\\()"'=>< are not allowed).
// 	operation can be "=", "<", "<=", ">", ">=", "CONTAINS" AND "EXISTS". operand
// 	can be a string (escaped with single quotes), number, date or time.

//	Examples:
//		  tm.event = 'NewBlock'               # new blocks
//		  tm.event = 'CompleteProposal'       # node got a complete proposal
//		  tm.event = 'Tx' AND tx.hash = 'XYZ' # single transaction
//		  tm.event = 'Tx' AND tx.height = 5   # all txs of the fifth block
//		  tx.height = 5                       # all txs of the fifth block
//
// For more information, see the /subscribe CometBFT RPC endpoint documentation
func QueryBlocks(ctx context.Context, rpcClient CometRPC, page, limit int, query, orderBy string) (*sdk.SearchBlocksResult, error) {
	resBlocks, err := rpcClient.BlockSearch(ctx, query, &page, &limit, orderBy)
	if err != nil {
		return nil, err
	}

	blocks, err := formatBlockResults(resBlocks.Blocks)
	if err != nil {
		return nil, err
	}

	result := NewSearchBlocksResult(int64(resBlocks.TotalCount), int64(len(blocks)), int64(page), int64(limit), blocks)

	return result, nil
}

// GetBlockByHeight gets block by height
func GetBlockByHeight(ctx context.Context, rpcClient CometRPC, height *int64) (*cmttypes.Block, error) {
	// header -> BlockchainInfo
	// header, tx -> Block
	// results -> BlockResults
	resBlock, err := rpcClient.Block(ctx, height)
	if err != nil {
		return nil, err
	}

	out, err := NewResponseResultBlock(resBlock)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("unable to create response block from comet result block: %v", resBlock)
	}

	return out, nil
}

// GetBlockByHash gets block by hash
func GetBlockByHash(ctx context.Context, rpcClient CometRPC, hashHexString string) (*cmttypes.Block, error) {
	hash, err := hex.DecodeString(hashHexString)
	if err != nil {
		return nil, err
	}

	resBlock, err := rpcClient.BlockByHash(ctx, hash)

	if err != nil {
		return nil, err
	} else if resBlock.Block == nil {
		return nil, fmt.Errorf("block not found with hash: %s", hashHexString)
	}
	out, err := NewResponseResultBlock(resBlock)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("unable to create response block from comet result block: %v", resBlock)
	}

	return out, nil
}
