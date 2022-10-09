package node

import (
	context "context"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ ServiceServer = queryServer{}
)

type queryServer struct {
	clientCtx client.Context
}

func NewQueryServer(clientCtx client.Context) ServiceServer {
	return queryServer{
		clientCtx: clientCtx,
	}
}

func (s queryServer) Config(ctx context.Context, req *ConfigRequest) (*ConfigResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &ConfigResponse{
		MinimumGasPrice: sdkCtx.MinGasPrices().String(),
	}, nil
}
