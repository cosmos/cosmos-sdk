package rpc_test

import (
	"context"
	"strconv"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/address"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

type IntegrationTestSuite struct {
	suite.Suite

	network network.NetworkI
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg, err := network.DefaultConfigWithAppConfig(network.MinimumAppConfig())

	s.NoError(err)

	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCLIQueryConn() {
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

func (s *IntegrationTestSuite) TestQueryABCIHeight() {
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
