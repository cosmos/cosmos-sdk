package log

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestLoggerCtx(t *testing.T) {
	logger := newSlogLogger("test", io.Discard, Config{})
	setVal := "bar"
	ctx := context.WithValue(context.Background(), "foo", setVal)
	logger = logger.Ctx(ctx)
	sLogger, ok := logger.(*verboseModeLogger)
	require.True(t, ok)

	val, ok := sLogger.ctx.Value("foo").(string)
	require.True(t, ok)
	require.Equal(t, setVal, val)
}

// this test ensures that when the With and WithContext methods are called,
// that the log wrapper is properly copied with all of its associated options
// otherwise, verbose mode will fail
func TestLoggerWith(t *testing.T) {
	logger := zerolog.New(&bytes.Buffer{})
	regularLevel := zerolog.WarnLevel
	verboseLevel := zerolog.InfoLevel
	filterWriter := &filterWriter{}
	wrapper := zeroLogWrapper{
		Logger:       &logger,
		regularLevel: regularLevel,
		verboseLevel: verboseLevel,
		filterWriter: filterWriter,
	}

	wrapper2 := wrapper.With("x", "y").(*zeroLogWrapper)
	if wrapper2.filterWriter != filterWriter {
		t.Fatalf("expected filterWriter to be copied, but it was not")
	}
	if wrapper2.regularLevel != regularLevel {
		t.Fatalf("expected regularLevel to be copied, but it was not")
	}
	if wrapper2.verboseLevel != verboseLevel {
		t.Fatalf("expected verboseLevel to be copied, but it was not")
	}
}
