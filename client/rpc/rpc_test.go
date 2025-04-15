package rpc_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/address"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
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

	testClient := testdata.NewQueryClient(s.network.Validators[0].ClientCtx)
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
		name                string
		reqHeight           int64
		ctxHeight           int64
		awaitMinChainHeight int64
		assertFn            func(t *testing.T, latestHeightAtQuery int64, resp abci.QueryResponse)
	}{
		{
			name:                "request height set",
			reqHeight:           2, // no proof when < 2
			ctxHeight:           1,
			awaitMinChainHeight: 3, // wait +1 block to be on the safe side
			assertFn: func(t *testing.T, _ int64, resp abci.QueryResponse) {
				t.Helper()
				assert.Equal(t, int64(2), resp.Height)
			},
		},
		{
			name:                "fallback to context height when request height is not set",
			reqHeight:           0,
			ctxHeight:           3,
			awaitMinChainHeight: 4, // wait +1 block to be on the safe side
			assertFn: func(t *testing.T, _ int64, resp abci.QueryResponse) {
				t.Helper()
				assert.Equal(t, int64(3), resp.Height)
			},
		},
		{
			name:                "with empty values, use latest height",
			reqHeight:           0,
			ctxHeight:           0,
			awaitMinChainHeight: 2, // no proof when < 2
			assertFn: func(t *testing.T, latestHeightAtQuery int64, resp abci.QueryResponse) {
				t.Helper()
				anyOf := []int64{latestHeightAtQuery, latestHeightAtQuery - 1}
				assert.Contains(t, anyOf, resp.Height)
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			currentHeight, err := s.network.WaitForHeight(tc.awaitMinChainHeight)
			s.Require().NoError(err)

			val := s.network.Validators[0]
			clientCtx := val.ClientCtx
			clientCtx = clientCtx.WithHeight(tc.ctxHeight)

			req := abci.QueryRequest{
				Path:   fmt.Sprintf("store/%s/key", banktypes.StoreKey),
				Height: tc.reqHeight,
				Data:   address.MustLengthPrefix(val.Address),
				Prove:  true,
			}

			res, err := clientCtx.QueryABCI(req)
			s.Require().NoError(err)
			tc.assertFn(s.T(), currentHeight, res)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
