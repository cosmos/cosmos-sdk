package simsx

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BlockTime read header block time from sdk context
func BlockTime(ctx context.Context) time.Time {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.BlockTime()
}
