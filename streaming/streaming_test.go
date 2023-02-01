package streaming

import (
	"fmt"
	"os"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PluginTestSuite struct {
	suite.Suite

	loggerCtx sdk.Context

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
	s.loggerCtx = sdk.NewContext(
		nil,
		tmproto.Header{Height: 1, Time: time.Now()},
		false,
		log.TestingLogger(),
	)

	path, err := os.Getwd()
	if err != nil {
		s.T().Fail()
	}
	s.workDir = path

	// test abci message types
	s.beginBlockReq = abci.RequestBeginBlock{
		Header:              tmproto.Header{Height: 1, Time: time.Now()},
		ByzantineValidators: []abci.Evidence{},
		Hash:                []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
		LastCommitInfo:      abci.LastCommitInfo{Round: 1, Votes: []abci.VoteInfo{}},
	}
	s.beginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{{Type: "testEventType1"}},
	}
	s.endBlockReq = abci.RequestEndBlock{Height: 1}
	s.endBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &abci.ConsensusParams{},
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

	// test storetypes kv pair types
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
		pluginVersion := "abci_v1"
		// to write data to files, replace stdout/stdout => file/file
		pluginPath := fmt.Sprintf("%s/plugins/abci/v1/examples/stdout/stdout", s.workDir)
		if err := os.Setenv(GetPluginEnvKey(pluginVersion), pluginPath); err != nil {
			t.Fail()
		}

		raw, err := NewStreamingPlugin(pluginVersion, "trace")
		require.NoError(t, err, "load", "streaming", "unexpected error")

		abciListener, ok := raw.(storetypes.ABCIListener)
		require.True(t, ok, "should pass type check")

		s.loggerCtx = s.loggerCtx.WithStreamingManager(storetypes.StreamingManager{
			AbciListeners: []storetypes.ABCIListener{abciListener},
			StopNodeOnErr: true,
		})

		for i := range [50]int{} {
			s.updateHeight(int64(i))

			err = abciListener.ListenBeginBlock(s.loggerCtx, s.beginBlockReq, s.beginBlockRes)
			assert.NoError(t, err, "ListenBeginBlock")

			err = abciListener.ListenEndBlock(s.loggerCtx, s.endBlockReq, s.endBlockRes)
			assert.NoError(t, err, "ListenEndBlock")

			for range [50]int{} {
				err = abciListener.ListenDeliverTx(s.loggerCtx, s.deliverTxReq, s.deliverTxRes)
				assert.NoError(t, err, "ListenDeliverTx")
			}

			err = abciListener.ListenCommit(s.loggerCtx, s.commitRes, s.changeSet)
			assert.NoError(t, err, "ListenCommit")
		}
	})
}

func (s *PluginTestSuite) updateHeight(n int64) {
	s.beginBlockReq.Header.Height = n
	s.endBlockReq.Height = n
	s.loggerCtx = s.loggerCtx.WithBlockHeight(n)
}
