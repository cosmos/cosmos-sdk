package baseapp

import (
	"context"
	"errors"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/log/v2"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

type countingLogger struct {
	debugCount int
	lastDebug  string
}

func (l *countingLogger) Info(_ string, _ ...any)                            {}
func (l *countingLogger) InfoContext(_ context.Context, _ string, _ ...any)  {}
func (l *countingLogger) Warn(_ string, _ ...any)                            {}
func (l *countingLogger) WarnContext(_ context.Context, _ string, _ ...any)  {}
func (l *countingLogger) Error(_ string, _ ...any)                           {}
func (l *countingLogger) ErrorContext(_ context.Context, _ string, _ ...any) {}
func (l *countingLogger) Debug(msg string, _ ...any) {
	l.debugCount++
	l.lastDebug = msg
}
func (l *countingLogger) DebugContext(_ context.Context, msg string, _ ...any) { l.Debug(msg) }
func (l *countingLogger) With(_ ...any) log.Logger                             { return l }
func (l *countingLogger) Impl() any                                            { return nil }

func TestWithGRPCBlockHeight_HeaderOk_NoTrailer(t *testing.T) {
	logger := &countingLogger{}
	app := NewBaseApp(t.Name(), logger, dbm.NewMemDB(), nil)

	var setTrailerCalls int
	setHeader := func(_ context.Context, md metadata.MD) error {
		require.Equal(t, []string{"10"}, md.Get(grpctypes.GRPCBlockHeightHeader))
		return nil
	}
	setTrailer := func(_ context.Context, _ metadata.MD) error {
		setTrailerCalls++
		return nil
	}

	resp, err := app.withGRPCBlockHeight(context.Background(), context.Background(), 10, setHeader, setTrailer, func() (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	require.Equal(t, "ok", resp)
	require.Equal(t, 0, setTrailerCalls)
	require.Equal(t, 0, logger.debugCount)
}

func TestWithGRPCBlockHeight_HeaderFails_TrailerFallback(t *testing.T) {
	logger := &countingLogger{}
	app := NewBaseApp(t.Name(), logger, dbm.NewMemDB(), nil)

	setHeaderErr := errors.New("transport: SendHeader called multiple times")
	var trailerMD metadata.MD
	setHeader := func(_ context.Context, _ metadata.MD) error { return setHeaderErr }
	setTrailer := func(_ context.Context, md metadata.MD) error {
		trailerMD = md
		return nil
	}

	resp, err := app.withGRPCBlockHeight(context.Background(), context.Background(), 12, setHeader, setTrailer, func() (any, error) {
		return 123, nil
	})
	require.NoError(t, err)
	require.Equal(t, 123, resp)
	require.Equal(t, []string{"12"}, trailerMD.Get(grpctypes.GRPCBlockHeightHeader))
	require.Equal(t, "failed to set gRPC header", logger.lastDebug)
	require.GreaterOrEqual(t, logger.debugCount, 1)
}
