package rpc

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tm "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// get the current blockchain height
func GetChainHeight(clientCtx client.Context) (int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return -1, err
	}

	status, err := node.Status(context.Background())
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// QueryBlocksByEvents performs a search for blocks for a given set of events
// via the Tendermint RPC. An event takes the form of:
// "{eventAttribute}.{attributeKey} = '{attributeValue}'". Each event is
// concatenated with an 'AND' operand. It returns a slice of Info object
// containing blocks and metadata. An error is returned if the query fails.
// If an empty string is provided it will order blocks by asc
func QueryBlocksByEvents(clientCtx client.Context, events []string, page, limit int, orderBy string) (*sdk.SearchBlocksResult, error) {
	if len(events) == 0 {
		return nil, errors.New("must declare at least one event to search")
	}

	if page <= 0 {
		return nil, errors.New("page must be greater than 0")
	}

	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	// XXX: implement ANY
	query := strings.Join(events, " AND ")

	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	// TODO: this may not always need to be proven
	// https://github.com/cosmos/cosmos-sdk/issues/6807
	resBlocks, err := node.BlockSearch(context.Background(), query, &page, &limit, orderBy)
	if err != nil {
		return nil, err
	}

	blocks, err := formatBlockResults(resBlocks.Blocks)
	if err != nil {
		return nil, err
	}

	result := sdk.NewSearchBlocksResult(int64(resBlocks.TotalCount), int64(len(blocks)), int64(page), int64(limit), blocks)

	return result, nil
}

// get block by height
func GetBlockByHeight(clientCtx client.Context, height *int64) (*tm.Block, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	// header -> BlockchainInfo
	// header, tx -> Block
	// results -> BlockResults
	resBlock, err := node.Block(context.Background(), height)
	if err != nil {
		return nil, err
	}

	out, err := mkBlockResult(resBlock)
	if err != nil {
		return out, err
	}

	return out, nil
}

func GetBlockByHash(clientCtx client.Context, hashHexString string) (*tm.Block, error) {
	hash, err := hex.DecodeString(hashHexString)
	if err != nil {
		return nil, err
	}

	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	resBlock, err := node.BlockByHash(context.Background(), hash)

	if err != nil || resBlock.Block == nil {
		return nil, fmt.Errorf("block not found with Hash: %s with Error: %s", hashHexString, err)
	}

	out, err := mkBlockResult(resBlock)
	if err != nil {
		return out, err
	}

	return out, nil
}

func mkBlockResult(resBlock *coretypes.ResultBlock) (*tm.Block, error) {
	return sdk.NewResponseResultBlock(resBlock, resBlock.Block.Time.Format(time.RFC3339)), nil
}

// formatBlockResults parses the indexed blocks into a slice of BlockResponse objects.
func formatBlockResults(resBlocks []*coretypes.ResultBlock) ([]*tm.Block, error) {
	var err error
	out := make([]*tm.Block, len(resBlocks))
	for i := range resBlocks {
		out[i], err = mkBlockResult(resBlocks[i])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}
