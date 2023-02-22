package streaming

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gogoproto/proto"

	storetypes "cosmossdk.io/store/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PluginTestSuite struct {
	suite.Suite

	loggerCtx MockContext

	workDir string

	beginBlockReq abci.RequestBeginBlock
	beginBlockRes abci.ResponseBeginBlock
	endBlockReq   abci.RequestEndBlock
	endBlockRes   abci.ResponseEndBlock
	deliverTxReq  abci.RequestDeliverTx
	deliverTxRes  abci.ResponseDeliverTx
	commitRes     abci.ResponseCommit

	changeSet []*storetypes.StoreKVPair
}

func (s *PluginTestSuite) SetupTest() {
	path, err := os.Getwd()
	if err != nil {
		s.T().Fail()
	}
	s.workDir = path

	pluginVersion := "abci"
	// to write data to files, replace stdout/stdout => file/file
	pluginPath := fmt.Sprintf("%s/abci/examples/stdout/stdout", s.workDir)
	if err := os.Setenv(GetPluginEnvKey(pluginVersion), pluginPath); err != nil {
		s.T().Fail()
	}

	raw, err := NewStreamingPlugin(pluginVersion, "trace")
	require.NoError(s.T(), err, "load", "streaming", "unexpected error")

	abciListener, ok := raw.(storetypes.ABCIListener)
	require.True(s.T(), ok, "should pass type check")

	header := tmproto.Header{Height: 1, Time: time.Now()}
	logger := log.NewNopLogger()
	streamingService := storetypes.StreamingManager{
		ABCIListeners: []storetypes.ABCIListener{abciListener},
		StopNodeOnErr: true,
	}
	s.loggerCtx = NewMockContext(header, logger, streamingService)

	// test abci message types
	s.beginBlockReq = abci.RequestBeginBlock{
		Header:              s.loggerCtx.BlockHeader(),
		ByzantineValidators: []abci.Misbehavior{},
		Hash:                []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
		LastCommitInfo:      abci.CommitInfo{Round: 1, Votes: []abci.VoteInfo{}},
	}
	s.beginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{{Type: "testEventType1"}},
	}
	s.endBlockReq = abci.RequestEndBlock{Height: s.loggerCtx.BlockHeight()}
	s.endBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &tmproto.ConsensusParams{},
		ValidatorUpdates:      []abci.ValidatorUpdate{},
	}
	s.deliverTxReq = abci.RequestDeliverTx{
		Tx: []byte{9, 8, 7, 6, 5, 4, 3, 2, 1},
	}
	s.deliverTxRes = abci.ResponseDeliverTx{
		Events:    []abci.Event{},
		Code:      1,
		Codespace: "mockCodeSpace",
		Data:      []byte{5, 6, 7, 8},
		GasUsed:   2,
		GasWanted: 3,
		Info:      "mockInfo",
		Log:       "mockLog",
	}
	s.commitRes = abci.ResponseCommit{}

	// test store kv pair types
	for range [2000]int{} {
		s.changeSet = append(s.changeSet, &storetypes.StoreKVPair{
			StoreKey: "mockStore",
			Delete:   false,
			Key:      []byte{1, 2, 3},
			Value:    []byte{3, 2, 1},
		})
	}
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

func (s *PluginTestSuite) TestABCIGRPCPlugin() {
	s.T().Run("Should successfully load streaming", func(t *testing.T) {

		abciListeners := s.loggerCtx.StreamingManager().ABCIListeners
		for _, abciListener := range abciListeners {
			for i := range [50]int{} {
				err := abciListener.ListenBeginBlock(s.loggerCtx, s.beginBlockReq, s.beginBlockRes)
				assert.NoError(t, err, "ListenBeginBlock")

				err = abciListener.ListenEndBlock(s.loggerCtx, s.endBlockReq, s.endBlockRes)
				assert.NoError(t, err, "ListenEndBlock")

				for range [50]int{} {
					err = abciListener.ListenDeliverTx(s.loggerCtx, s.deliverTxReq, s.deliverTxRes)
					assert.NoError(t, err, "ListenDeliverTx")
				}

				err = abciListener.ListenCommit(s.loggerCtx, s.commitRes, s.changeSet)
				assert.NoError(t, err, "ListenCommit")

				s.updateHeight(int64(i + 1))
			}
		}
	})
}

func (s *PluginTestSuite) updateHeight(n int64) {
	header := s.loggerCtx.BlockHeader()
	header.Height = n
	s.loggerCtx = NewMockContext(header, s.loggerCtx.Logger(), s.loggerCtx.StreamingManager())
}

var _ context.Context = MockContext{}
var _ storetypes.Context = MockContext{}

type MockContext struct {
	baseCtx          context.Context
	header           tmproto.Header
	logger           log.Logger
	streamingManager storetypes.StreamingManager
}

func (m MockContext) BlockHeight() int64                            { return m.header.Height }
func (m MockContext) Logger() log.Logger                            { return m.logger }
func (m MockContext) StreamingManager() storetypes.StreamingManager { return m.streamingManager }

func (m MockContext) BlockHeader() tmproto.Header {
	msg := proto.Clone(&m.header).(*tmproto.Header)
	return *msg
}

func NewMockContext(header tmproto.Header, logger log.Logger, sm storetypes.StreamingManager) MockContext {
	header.Time = header.Time.UTC()
	return MockContext{
		baseCtx:          context.Background(),
		header:           header,
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
