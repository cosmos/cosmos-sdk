package watchers

import (
	"context"
	"time"

	"cosmossdk.io/tools/cosmovisor/internal/checkers"
)

func NewHeightWatcher(ctx context.Context, checker checkers.HeightChecker, pollInterval time.Duration) Watcher[uint64] {
	return NewPollWatcher[uint64](ctx, checker.GetLatestBlockHeight, pollInterval)
}
