package block

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

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

	result := sdk.NewSearchBlocksResult(uint64(resBlocks.TotalCount), uint64(len(blocks)), uint64(page), uint64(limit), blocks)

	return result, nil
}

// get block by height
func GetBlockByHeight(clientCtx client.Context, height *int64) (*sdk.BlockResponse, error) {
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

func GetBlockByHash(clientCtx client.Context, hashHexString string) (*sdk.BlockResponse, error) {

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
	if err != nil {
		return nil, err
	}

	out, err := mkBlockResult(resBlock)
	if err != nil {
		return out, err
	}

	return out, nil
}

func mkBlockResult(resBlock *coretypes.ResultBlock) (*sdk.BlockResponse, error) {
	return sdk.NewResponseResultBlock(resBlock, resBlock.Block.Time.Format(time.RFC3339)), nil
}

// formatBlockResults parses the indexed blocks into a slice of BlockResponse objects.
func formatBlockResults(resBlocks []*coretypes.ResultBlock) ([]*sdk.BlockResponse, error) {
	var err error
	out := make([]*sdk.BlockResponse, len(resBlocks))
	for i := range resBlocks {
		out[i], err = mkBlockResult(resBlocks[i])
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}
