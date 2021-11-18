package tmservice

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/tendermint/tendermint/rpc/coretypes"
)

func getBlock(ctx context.Context, clientCtx client.Context, height *int64) (*coretypes.ResultBlock, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	return node.Block(ctx, height)
}
