package streaming

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/log"
	coretesting "cosmossdk.io/core/testing"
)

type PluginTestSuite struct {
	suite.Suite

	loggerCtx MockContext

	workDir string

	deliverBlockrequest ListenDeliverBlockRequest
	stateChangeRequest  ListenStateChangesRequest

	changeSet []*StoreKVPair
}

func (s *PluginTestSuite) SetupTest() {
	if runtime.GOOS != "linux" {
		s.T().Skip("only run on linux")
	}

	path, err := os.Getwd()
	if err != nil {
		s.T().Fail()
	}
	s.workDir = path

	pluginVersion := defaultPlugin
	// to write data to files, replace stdout/stdout => file/file
	pluginPath := fmt.Sprintf("%s/abci/examples/stdout/stdout", s.workDir)
	if err := os.Setenv(GetPluginEnvKey(pluginVersion), pluginPath); err != nil {
		s.T().Fail()
	}

	raw, err := NewStreamingPlugin(pluginVersion, "trace")
	require.NoError(s.T(), err, "load", "streaming", "unexpected error")

	abciListener, ok := raw.(Listener)
	require.True(s.T(), ok, "should pass type check")

	logger := coretesting.NewNopLogger()
	streamingService := Manager{
		Listeners:     []Listener{abciListener},
		StopNodeOnErr: true,
	}
	s.loggerCtx = NewMockContext(1, logger, streamingService)

	// test abci message types

	s.deliverBlockrequest = ListenDeliverBlockRequest{
		BlockHeight: s.loggerCtx.BlockHeight(),
		Txs:         [][]byte{{1, 2, 3, 4, 5, 6, 7, 8, 9}},
		Events:      []*Event{},
	}
	s.stateChangeRequest = ListenStateChangesRequest{}

	key := []byte("mockStore")
	key = append(key, 1, 2, 3)
	// test store kv pair types
	for range [2000]int{} {
		s.changeSet = append(s.changeSet, &StoreKVPair{
			Key:   key,
			Value: []byte{3, 2, 1},
		})
	}
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

func (s *PluginTestSuite) TestABCIGRPCPlugin() {
	s.T().Run("Should successfully load streaming", func(t *testing.T) {
		abciListeners := s.loggerCtx.StreamingManager().Listeners
		for _, abciListener := range abciListeners {
			for i := range [50]int{} {

				err := abciListener.ListenDeliverBlock(s.loggerCtx, s.deliverBlockrequest)
				assert.NoError(t, err, "ListenEndBlock")

				err = abciListener.ListenStateChanges(s.loggerCtx, s.changeSet)
				assert.NoError(t, err, "ListenCommit")

				s.updateHeight(int64(i + 1))
			}
		}
	})
}

func (s *PluginTestSuite) updateHeight(n int64) {
	s.loggerCtx = NewMockContext(n, s.loggerCtx.Logger(), s.loggerCtx.StreamingManager())
}

var (
	_ context.Context = MockContext{}
	_ Context         = MockContext{}
)

type MockContext struct {
	baseCtx          context.Context
	height           int64
	logger           log.Logger
	streamingManager Manager
}

func (m MockContext) BlockHeight() int64        { return m.height }
func (m MockContext) Logger() log.Logger        { return m.logger }
func (m MockContext) StreamingManager() Manager { return m.streamingManager }

func NewMockContext(height int64, logger log.Logger, sm Manager) MockContext {
	return MockContext{
		baseCtx:          context.Background(),
		height:           height,
		logger:           logger,
		streamingManager: sm,
	}
}

func (m MockContext) Deadline() (deadline time.Time, ok bool) {
	return m.baseCtx.Deadline()
}

func (m MockContext) Done() <-chan struct{} {
	return m.baseCtx.Done()
}

func (m MockContext) Err() error {
	return m.baseCtx.Err()
}

func (m MockContext) Value(key any) any {
	return m.baseCtx.Value(key)
}
