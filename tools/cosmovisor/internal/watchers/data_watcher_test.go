package watchers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestData struct {
	X int    `json:"x"`
	Y string `json:"y"`
}

func TestDataWatcher(t *testing.T) {
	dir, err := os.MkdirTemp("", "watcher")
	require.NoError(t, err)
	filename := filepath.Join(dir, "testfile.json")

	ctx, cancel := context.WithCancel(context.Background())
	pollWatcher := NewFilePollWatcher(ctx, filename, time.Millisecond*100)
	dataWatcher := NewDataWatcher[TestData](ctx, pollWatcher, func(contents []byte) (TestData, error) {
		var data TestData
		err := json.Unmarshal(contents, &data)
		return data, err
	})

	expectedContent := TestData{
		X: 10,
		Y: "testtesttest",
	}
	go func() {
		// write some dummy data to the file
		time.Sleep(time.Second)
		err = os.WriteFile(filename, []byte("unexpected content - should be ignored"), 0644)
		require.NoError(t, err)

		// write the expected content to the file
		time.Sleep(time.Second)
		bz, err := json.Marshal(expectedContent)
		require.NoError(t, err)
		err = os.WriteFile(filename, bz, 0644)
		require.NoError(t, err)

		// wait a bit to ensure the watcher has time to pick up the change
		// then cancel the context
		time.Sleep(time.Second)
		cancel()
	}()

	var actualContext *TestData

	// we check all the channels in a function which we'll return from whenever
	// a channel is closed or we get the done signal
	func() {
		for {
			select {
			case content, ok := <-dataWatcher.Updated():
				if !ok {
					return
				}
				actualContext = &content
			case err, ok := <-dataWatcher.Errors():
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
	require.Equal(t, expectedContent, *actualContext)

	// check that all the channels are closed
	_, open := <-dataWatcher.Updated()
	require.False(t, open)
	_, open = <-dataWatcher.Errors()
	require.False(t, open)

}
