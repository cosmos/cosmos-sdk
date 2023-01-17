package block

import (
	"context"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

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
	return sdk.NewResponseResultBlock(resBlock), nil
}
