package watchers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
)

func TestPollWatcher(t *testing.T) {
	dir, err := os.MkdirTemp("", "watcher")
	require.NoError(t, err)
	filename := filepath.Join(dir, "testfile")

	ctx, cancel := context.WithCancel(context.Background())
	eh := DebugLoggerErrorHandler(log.NewTestLogger(t))
	watcher := NewFilePollWatcher(ctx, eh, filename, time.Millisecond*100)
	expectedContent := []byte("test")
	go func() {
		// write some dummy data to the file
		time.Sleep(time.Second)
		err = os.WriteFile(filename, []byte("unexpected content - should be updated later"), 0o644)
		require.NoError(t, err)

		// write the expected content to the file
		time.Sleep(time.Second)
		err := os.WriteFile(filename, expectedContent, 0o644)
		require.NoError(t, err)

		// wait a bit to ensure the watcher has time to pick up the change
		// then cancel the context
		time.Sleep(time.Second)
		cancel()
	}()

	var actualContent []byte
	// we check all the channels in a function which we'll return from whenever
	// a channel is closed or we get the done signal
	func() {
		for {
			select {
			case bz, ok := <-watcher.Updated():
				if !ok {
					return
				}
				actualContent = bz
			case <-ctx.Done():
				return
			}
		}
	}()

	// check we have the expected context
	require.Equal(t, expectedContent, actualContent)

	// check that all the channels are closed
	_, open := <-watcher.Updated()
	require.False(t, open)
}
