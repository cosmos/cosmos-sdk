package watchers

import (
	"context"
	"time"

	"cosmossdk.io/log"
)

type HeightChecker interface {
	GetLatestBlockHeight() (uint64, error)
}

type HeightWatcher struct {
	*PollWatcher[uint64]
	checker     HeightChecker
	onGetHeight func(uint64) error
}

func NewHeightWatcher(ctx context.Context, logger log.Logger, checker HeightChecker, pollInterval time.Duration, onGetHeight func(uint64) error) *HeightWatcher {
	watcher := &HeightWatcher{
		checker:     checker,
		onGetHeight: onGetHeight,
	}
	watcher.PollWatcher = NewPollWatcher[uint64](ctx, logger, func() (uint64, error) {
		return watcher.ReadNow()

	}, pollInterval)
	return watcher
}

func (h HeightWatcher) ReadNow() (uint64, error) {
	height, err := h.checker.GetLatestBlockHeight()
	if err != nil {
		return 0, err
	}
	err = h.onGetHeight(height)
	return height, err
}
