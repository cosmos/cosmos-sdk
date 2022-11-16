package streaming

import (
	"fmt"
	store "github.com/cosmos/cosmos-sdk/store/types"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"os"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
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

	changeSet []*store.StoreKVPair
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
	s.changeSet = []*types.StoreKVPair{
		{
			StoreKey: "mockStore",
			Delete:   false,
			Key:      []byte{1, 2, 3},
			Value:    []byte{3, 2, 1},
		},
		{
			StoreKey: "mockStore",
			Delete:   false,
			Key:      []byte{3, 4, 5},
			Value:    []byte{5, 4, 3},
		},
	}
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

func (s *PluginTestSuite) TestABCIGRPCPlugin() {
	s.T().Run("Should successfully load streaming", func(t *testing.T) {
		pluginVersion := "abci_v1"
		//pluginPath := fmt.Sprintf("%s/plugins/abci/v1/examples/plugin-go/stdout", s.workDir)
		pluginPath := fmt.Sprintf("python3 %s/plugins/abci/v1/examples/plugin-python/file.py", s.workDir)
		//pluginPath := fmt.Sprintf("python3 %s/plugins/abci/v1/examples/plugin-python/kafka.py", s.workDir)
		if err := os.Setenv(GetPluginEnvKey(pluginVersion), pluginPath); err != nil {
			t.Fail()
		}

		raw, err := NewStreamingPlugin(pluginVersion, "trace")
		require.NoError(t, err, "load", "streaming", "unexpected error")

		abciListener, ok := raw.(baseapp.ABCIListener)
		require.True(t, ok, "should pass type check")

		err = abciListener.ListenBeginBlock(s.loggerCtx, s.beginBlockReq, s.beginBlockRes)
		assert.NoError(t, err, "ListenBeginBlock")

		err = abciListener.ListenEndBlock(s.loggerCtx, s.endBlockReq, s.endBlockRes)
		assert.NoError(t, err, "ListenEndBlock")

		err = abciListener.ListenDeliverTx(s.loggerCtx, s.deliverTxReq, s.deliverTxRes)
		assert.NoError(t, err, "ListenDeliverTx")
		err = abciListener.ListenDeliverTx(s.loggerCtx, s.deliverTxReq, s.deliverTxRes)
		assert.NoError(t, err, "ListenDeliverTx")

		// streaming services can choose not to implement store listening
		err = abciListener.ListenCommit(s.loggerCtx, s.commitRes, s.changeSet)
		assert.NoError(t, err, "ListenCommit")
	})
}
