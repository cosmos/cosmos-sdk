package baseapp

import (
	"context"
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/log/v2"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

type countingLogger struct {
	debugCount           int
	failedSetHeaderCount int
	lastDebug            string
}

func (l *countingLogger) Info(_ string, _ ...any)                            {}
func (l *countingLogger) InfoContext(_ context.Context, _ string, _ ...any)  {}
func (l *countingLogger) Warn(_ string, _ ...any)                            {}
func (l *countingLogger) WarnContext(_ context.Context, _ string, _ ...any)  {}
func (l *countingLogger) Error(_ string, _ ...any)                           {}
func (l *countingLogger) ErrorContext(_ context.Context, _ string, _ ...any) {}
func (l *countingLogger) Debug(msg string, _ ...any) {
	l.debugCount++
	if msg == "failed to set gRPC header" {
		l.failedSetHeaderCount++
	}
	l.lastDebug = msg
}
func (l *countingLogger) DebugContext(_ context.Context, msg string, _ ...any) { l.Debug(msg) }
func (l *countingLogger) With(_ ...any) log.Logger                             { return l }
func (l *countingLogger) Impl() any                                            { return nil }

type fakeServerTransportStream struct {
	method string

	headersSent bool
	headerMD    metadata.MD
	trailerMD   metadata.MD
}

func (s *fakeServerTransportStream) Method() string { return s.method }

func (s *fakeServerTransportStream) SetHeader(md metadata.MD) error {
	if s.headersSent {
		return errors.New("transport: SendHeader called multiple times")
	}
	s.headerMD = metadata.Join(s.headerMD, md)
	return nil
}

func (s *fakeServerTransportStream) SendHeader(md metadata.MD) error {
	if s.headersSent {
		return errors.New("transport: SendHeader called multiple times")
	}
	s.headersSent = true
	s.headerMD = metadata.Join(s.headerMD, md)
	return nil
}

func (s *fakeServerTransportStream) SetTrailer(md metadata.MD) error {
	s.trailerMD = metadata.Join(s.trailerMD, md)
	return nil
}

func setupBaseAppForGRPCQueryTests(t *testing.T, logger log.Logger) *BaseApp {
	t.Helper()

	app := NewBaseApp(t.Name(), logger, dbm.NewMemDB(), nil)
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	return app
}

func TestGRPCQueryInterceptor_BlockHeightHeaderOk_NoTrailer(t *testing.T) {
	testCases := []struct {
		name                  string
		preSendHeaders        bool
		expectedResp          any
		expectedHeaderHeight  []string
		expectedTrailerHeight []string
		expectedMinFailCount  int
		expectedLastDebug     string
	}{
		{
			name:                  "header_ok",
			preSendHeaders:        false,
			expectedResp:          "ok",
			expectedHeaderHeight:  []string{"1"},
			expectedTrailerHeight: nil,
			expectedMinFailCount:  0,
			expectedLastDebug:     "",
		},
		{
			name:                  "header_already_sent_trailer_fallback",
			preSendHeaders:        true,
			expectedResp:          123,
			expectedHeaderHeight:  nil,
			expectedTrailerHeight: []string{"1"},
			expectedMinFailCount:  1,
			expectedLastDebug:     "failed to set gRPC header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &countingLogger{}
			app := setupBaseAppForGRPCQueryTests(t, logger)
			interceptor := app.grpcQueryInterceptor(false)

			stream := &fakeServerTransportStream{method: "/test.TestService/TestMethod"}
			grpcCtx := grpc.NewContextWithServerTransportStream(context.Background(), stream)
			grpcCtx = metadata.NewIncomingContext(grpcCtx, metadata.MD{})

			if tc.preSendHeaders {
				require.NoError(t, grpc.SendHeader(grpcCtx, metadata.Pairs("pre", "sent")))
			}

			resp, err := interceptor(grpcCtx, struct{}{}, &grpc.UnaryServerInfo{FullMethod: stream.method}, func(_ context.Context, _ any) (any, error) {
				return tc.expectedResp, nil
			})
			require.NoError(t, err)
			require.Equal(t, tc.expectedResp, resp)

			require.Equal(t, tc.expectedHeaderHeight, stream.headerMD.Get(grpctypes.GRPCBlockHeightHeader))
			require.Equal(t, tc.expectedTrailerHeight, stream.trailerMD.Get(grpctypes.GRPCBlockHeightHeader))
			require.GreaterOrEqual(t, logger.failedSetHeaderCount, tc.expectedMinFailCount)
			if tc.expectedLastDebug != "" {
				require.Equal(t, tc.expectedLastDebug, logger.lastDebug)
			}
		})
	}
}
