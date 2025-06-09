package watchers

import (
	"context"
	"fmt"
	"os"
	"time"

	"cosmossdk.io/log"
)

func NewFilePollWatcher(ctx context.Context, logger log.Logger, filename string, pollInterval time.Duration) Watcher[[]byte] {
	stat, err := os.Stat(filename)
	var lastModTime time.Time
	if err == nil {
		lastModTime = stat.ModTime()
	}
	check := func() ([]byte, error) {
		stat, err := os.Stat(filename)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to stat file %s: %w", filename, err)
			}
		} else {
			modTime := stat.ModTime()
			if stat.Size() > 0 && !modTime.Equal(lastModTime) {
				lastModTime = modTime
				bz, err := os.ReadFile(filename)
				if err != nil {
					return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
				} else {
					return bz, nil
				}
			}
		}
		return nil, os.ErrNotExist
	}
	return NewPollWatcher[[]byte](ctx, logger, check, pollInterval)
}
