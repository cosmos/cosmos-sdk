package simsx

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BlockTime read header block time from sdk context or sims context key if not present
func BlockTime(ctx context.Context) time.Time {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.BlockTime()
}
