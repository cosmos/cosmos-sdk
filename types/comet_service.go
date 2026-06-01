package types

import (
	"context"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
)

// ProvideCometInfoService returns a comet.BlockInfoService backed by the SDK context.
func ProvideCometInfoService() comet.BlockInfoService {
	return cometInfoService{}
}

var _ comet.BlockInfoService = cometInfoService{}

type cometInfoService struct{}

func (cometInfoService) GetCometBlockInfo(ctx context.Context) comet.BlockInfo {
	return UnwrapSDKContext(ctx).CometInfo()
}

var _ header.Service = headerInfoService{}

type headerInfoService struct{}

func (headerInfoService) GetHeaderInfo(ctx context.Context) header.Info {
	return UnwrapSDKContext(ctx).HeaderInfo()
}
