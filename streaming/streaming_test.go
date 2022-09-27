package streaming

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/streaming/plugins/abci"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PluginTestSuite struct {
	suite.Suite

	emptyCtx  sdk.Context
	loggerCtx sdk.Context

	workDir string
}

func (s *PluginTestSuite) SetupTest() {
	s.emptyCtx = sdk.Context{}
	s.loggerCtx = s.emptyCtx.
		WithLogger(log.TestingLogger()).
		WithBlockHeight(1).
		WithBlockHeader(tmproto.Header{Time: time.Now()})

	path, err := os.Getwd()
	if err != nil {
		s.T().Fail()
	}
	s.workDir = path
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

func (s *PluginTestSuite) TestABCIGRPCPlugin() {
	s.T().Run("Should successfully load streaming", func(t *testing.T) {
		plugin := "abci"
		pluginPath := fmt.Sprintf("%s/plugins/%s/examples/plugin-go/stdout", s.workDir, "abci")
		if err := os.Setenv(GetPluginEnvKey(plugin), pluginPath); err != nil {
			t.Fail()
		}

		raw, err := NewStreamingPlugin(plugin)
		require.NoError(t, err, "load", "streaming", "unexpected error")

		listener, ok := raw.(abci.Listener)
		require.True(t, ok, "should pass type check")

		err = listener.ListenBeginBlock(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")

		err = listener.ListenEndBlock(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")

		err = listener.ListenDeliverTx(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")

		err = listener.ListenStoreKVPair(s.loggerCtx.BlockHeight(), []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")
	})
}
