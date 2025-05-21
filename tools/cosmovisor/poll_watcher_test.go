package cosmovisor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPollWatcher(t *testing.T) {
	dir, err := os.MkdirTemp("", "watcher")
	require.NoError(t, err)
	filename := filepath.Join(dir, "testfile")

	ctx, cancel := context.WithCancel(context.Background())
	watcher := NewPollWatcher(ctx, filename, time.Millisecond*100)
	expectedContext := []byte("test")
	go func() {
		time.Sleep(time.Second)
		err := os.WriteFile(filename, expectedContext, 0644)
		require.NoError(t, err)
		time.Sleep(time.Second)
		cancel()
	}()
	var actualContext []byte

	// we check all the channels in a function which we'll return from whenever
	// a channel is closed or we get the done signal
	func() {
		for {
			select {
			case bz, ok := <-watcher.Updated():
				if !ok {
					return
				}
				actualContext = bz
			case err, ok := <-watcher.Errors():
				if !ok {
					return
				}
				require.NoError(t, err)
			case <-ctx.Done():
				return
			}
		}
	}()

	// check we have the expected context
	require.Equal(t, expectedContext, actualContext)

	// check that all the channels are closed
	_, open := <-watcher.Updated()
	require.False(t, open)
	_, open = <-watcher.Errors()
	require.False(t, open)

}
