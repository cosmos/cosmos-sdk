package tmservice

import (
	"context"

	"github.com/tendermint/tendermint/rpc/coretypes"

	"github.com/cosmos/cosmos-sdk/client"
)

func getNodeStatus(ctx context.Context, clientCtx client.Context) (*coretypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &coretypes.ResultStatus{}, err
	}
	return node.Status(ctx)
}
