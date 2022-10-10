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
	typesCfg := sdk.GetConfig()
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &ConfigResponse{
		MinimumGasPrice:     sdkCtx.MinGasPrices().String(),
		Bech32AccountAddr:   typesCfg.GetBech32AccountAddrPrefix(),
		Bech32ValidatorAddr: typesCfg.GetBech32ValidatorAddrPrefix(),
		Bech32ConsensusAddr: typesCfg.GetBech32ConsensusAddrPrefix(),
		Bech32AccountPub:    typesCfg.GetBech32AccountPubPrefix(),
		Bech32ValidatorPub:  typesCfg.GetBech32ValidatorPubPrefix(),
		Bech32ConsensusPub:  typesCfg.GetBech32ConsensusPubPrefix(),
	}, nil
}
