package sdk

import (
	"context"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmos/cosmos-sdk/client/rpc"
)

func (c Client) GetNodeInfo(ctx context.Context) (rpc.NodeInfoResponse, error) {
	status, err := c.clientCtx.Client.Status(ctx)
	if err != nil {
		return rpc.NodeInfoResponse{}, err
	}
	return rpc.NodeInfoResponse{
		DefaultNodeInfo: status.NodeInfo,
		// NOTE(fdymylja): I doubt this is correct as we could run a rosetta
		// 'node' with version v0.40.1, and query a v0.40.2 node successfully
		ApplicationVersion: version.NewInfo(),
	}, nil
}
