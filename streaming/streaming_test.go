package streaming

import (
	"fmt"
	"os"
	"testing"
	"time"

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

	emptyCtx  sdk.Context
	loggerCtx sdk.Context

	workDir string
}

func (s *PluginTestSuite) SetupTest() {
	s.emptyCtx = sdk.Context{}
	s.loggerCtx = s.emptyCtx.
		WithLogger(log.TestingLogger()).
		WithBlockHeader(tmproto.Header{Height: 1, Time: time.Now()})

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
		pluginVersion := "grpc_abci_v1"
		pluginPath := fmt.Sprintf("%s/plugins/abci/%s/examples/plugin-go/stdout", s.workDir, pluginVersion)
		//pluginPath := fmt.Sprintf("python3 %s/plugins/abci/%s/examples/plugin-python/file.py", s.workDir, pluginVersion)
		//pluginPath := fmt.Sprintf("python3 %s/plugins/abci/%s/examples/plugin-python/kafka.py", s.workDir, pluginVersion)
		if err := os.Setenv(GetPluginEnvKey(pluginVersion), pluginPath); err != nil {
			t.Fail()
		}

		raw, err := NewStreamingPlugin(pluginVersion, "trace")
		require.NoError(t, err, "load", "streaming", "unexpected error")

		listener, ok := raw.(baseapp.ABCIListener)
		require.True(t, ok, "should pass type check")

		err = listener.ListenBeginBlock(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")

		err = listener.ListenEndBlock(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")

		err = listener.ListenDeliverTx(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")
		err = listener.ListenDeliverTx(s.loggerCtx.BlockHeight(), []byte{1, 2, 3}, []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")

		err = listener.ListenStoreKVPair(s.loggerCtx.BlockHeight(), []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")
		err = listener.ListenStoreKVPair(s.loggerCtx.BlockHeight(), []byte{1, 2, 3})
		assert.NoError(t, err, "ListenBeginBlock")
	})
}
