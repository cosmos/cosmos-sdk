package runtime

import (
	"context"

	corecomet "cosmossdk.io/core/comet"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ corecomet.Service = &ContextAwareCometInfoService{}

type ContextAwareCometInfoService struct{}

func (c ContextAwareCometInfoService) CometInfo(ctx context.Context) corecomet.Info {
	return sdk.UnwrapSDKContext(ctx).CometInfo()
}
