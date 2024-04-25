package network_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/simapp/network"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/auth"
	_ "cosmossdk.io/x/auth/tx/config"
	_ "cosmossdk.io/x/bank"
	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/address"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	_ "github.com/cosmos/cosmos-sdk/x/genutil"
)

// https://github.com/improbable-eng/grpc-web/blob/master/go/grpcweb/wrapper_test.go used as a reference
// to setup grpcRequest config.

const grpcWebContentType = "application/grpc-web"

type NetworkTestSuite struct {
	suite.Suite

	cfg      network.Config
	network  network.NetworkI
	protoCdc *codec.ProtoCodec
}

func (s *NetworkTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg, err := network.DefaultConfigWithAppConfig(network.MinimumAppConfig())
	fmt.Println("DefaultConfigWithAppConfig", err)

	s.NoError(err)
	cfg.NumValidators = 1
	s.cfg = cfg

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(2)
	s.Require().NoError(err)

	s.protoCdc = codec.NewProtoCodec(s.cfg.InterfaceRegistry)
}

func (s *NetworkTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}

func (s *NetworkTestSuite) TestCLIQueryConn() {
	s.T().Skip("data race in comet is causing this to fail")
	var header metadata.MD

	testClient := testdata.NewQueryClient(s.network.GetValidators()[0].GetClientCtx())
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"}, grpc.Header(&header))
	s.NoError(err)

	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	height, err := strconv.Atoi(blockHeight[0])
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(height, 1) // at least the 1st block

	s.Equal("hello", res.Message)
}

func (s *NetworkTestSuite) TestQueryABCIHeight() {
	testCases := []struct {
		name      string
		reqHeight int64
		ctxHeight int64
		expHeight int64
	}{
		{
			name:      "non zero request height",
			reqHeight: 3,
			ctxHeight: 1, // query at height 1 or 2 would cause an error
			expHeight: 3,
		},
		{
			name:      "empty request height - use context height",
			reqHeight: 0,
			ctxHeight: 3,
			expHeight: 3,
		},
		{
			name:      "empty request height and context height - use latest height",
			reqHeight: 0,
			ctxHeight: 0,
			expHeight: 4,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.network.WaitForHeight(tc.expHeight)
			s.Require().NoError(err)

			val := s.network.GetValidators()[0]

			clientCtx := val.GetClientCtx()
			clientCtx = clientCtx.WithHeight(tc.ctxHeight)

			req := abci.RequestQuery{
				Path:   "store/bank/key",
				Height: tc.reqHeight,
				Data:   address.MustLengthPrefix(val.GetAddress()),
				Prove:  true,
			}

			res, err := clientCtx.QueryABCI(req)
			s.Require().NoError(err)

			s.Require().Equal(tc.expHeight, res.Height)
		})
	}
}

func TestStatusCommand(t *testing.T) {
	t.Skip() // https://github.com/cosmos/cosmos-sdk/issues/17446

	cfg, err := network.DefaultConfigWithAppConfig(network.MinimumAppConfig())
	require.NoError(t, err)

	network, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	require.NoError(t, network.WaitForNextBlock())

	val0 := network.GetValidators()[0]
	cmd := server.StatusCommand()

	out, err := clitestutil.ExecTestCLICmd(val0.GetClientCtx(), cmd, []string{})
	require.NoError(t, err)

	// Make sure the output has the validator moniker.
	require.Contains(t, out.String(), fmt.Sprintf("\"moniker\":\"%s\"", val0.GetMoniker()))
}
