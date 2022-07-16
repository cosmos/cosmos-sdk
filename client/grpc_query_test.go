//go:build norace
// +build norace

package client_test

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type testcase struct {
	clientContextHeight int64
	grpcHeight          int64
	expectedHeight      int64
}

const (
	// if clientContextHeight or grpcHeight is set to this flag,
	// the test assumes that the respective height is not provided.
	heightNotSetFlag = int64(-1)
	// given the current block time, this should never be reached by the time
	// a test is run.
	invalidBeyondLatestHeight = 1_000_000_000
	// if this flag is set to expectedHeight, an error is assumed.
	errorHeightFlag = int64(-2)
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), network.DefaultConfig())
	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(3)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

//func (s *IntegrationTestSuite) TestGRPCQuery() {
//	val0 := s.network.Validators[0]
//
//	// gRPC query to test service should work
//	testClient := testdata.NewQueryClient(val0.ClientCtx)
//	testRes, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
//	s.Require().NoError(err)
//	s.Require().Equal("hello", testRes.Message)
//
//	// gRPC query to bank service should work
//	denom := fmt.Sprintf("%stoken", val0.Moniker)
//	bankClient := banktypes.NewQueryClient(val0.ClientCtx)
//	var header metadata.MD
//	bankRes, err := bankClient.Balance(
//		context.Background(),
//		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
//		grpc.Header(&header), // Also fetch grpc header
//	)
//	s.Require().NoError(err)
//	s.Require().Equal(
//		sdk.NewCoin(denom, s.network.Config.AccountTokens),
//		*bankRes.GetBalance(),
//	)
//	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
//	s.Require().NotEmpty(blockHeight[0]) // Should contain the block height
//
//	// Request metadata should work
//	val0.ClientCtx = val0.ClientCtx.WithHeight(1) // We set clientCtx to height 1
//	bankClient = banktypes.NewQueryClient(val0.ClientCtx)
//	bankRes, err = bankClient.Balance(
//		context.Background(),
//		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
//		grpc.Header(&header),
//	)
//	blockHeight = header.Get(grpctypes.GRPCBlockHeightHeader)
//	s.Require().Equal([]string{"1"}, blockHeight)
//}

func (s *IntegrationTestSuite) TestGRPCQuery_TestService() {
	val0 := s.network.Validators[0]

	// gRPC query to test service should work
	testClient := testdata.NewQueryClient(val0.ClientCtx)
	testRes, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestGRPCConcurrency() {
	val0 := s.network.Validators[0]
	clientCtx := val0.ClientCtx
	clientCtx.GRPCConcurrency = true
	in := &testdata.EchoRequest{Message: "hello"}
	out := &testdata.EchoResponse{}
	err := clientCtx.Invoke(context.Background(), "/testdata.Query/Echo", in, out)
	s.Require().NoError(err)
	s.Require().Equal("hello", out.Message)

	clientCtx.GRPCConcurrency = false
	err = clientCtx.Invoke(context.Background(), "/testdata.Query/Echo", in, out)
	s.Require().NoError(err)
	s.Require().Equal("hello", out.Message)
}

func (s *IntegrationTestSuite) TestGRPCQuery_BankService_VariousInputs() {
	val0 := s.network.Validators[0]

	const method = "/cosmos.bank.v1beta1.Query/Balance"

	testcases := map[string]testcase{
		"clientContextHeight 1; grpcHeight not set - clientContextHeight selected": {
			clientContextHeight: 1, // chosen
			grpcHeight:          heightNotSetFlag,
			expectedHeight:      1,
		},
		"clientContextHeight not set; grpcHeight is 2 - grpcHeight is chosen": {
			clientContextHeight: heightNotSetFlag,
			grpcHeight:          2, // chosen
			expectedHeight:      2,
		},
		"both not set - 0 returned": {
			clientContextHeight: heightNotSetFlag,
			grpcHeight:          heightNotSetFlag,
			expectedHeight:      3, // latest height
		},
		"clientContextHeight 3; grpcHeight is 0 - grpcHeight is chosen": {
			clientContextHeight: 1,
			grpcHeight:          0, // chosen
			expectedHeight:      3, // latest height
		},
		"clientContextHeight 3; grpcHeight is 3 - 3 is returned": {
			clientContextHeight: 3,
			grpcHeight:          3,
			expectedHeight:      3,
		},
		"clientContextHeight is 1_000_000_000; grpcHeight is 1_000_000_000 - requested beyond latest height - error": {
			clientContextHeight: invalidBeyondLatestHeight,
			grpcHeight:          invalidBeyondLatestHeight,
			expectedHeight:      errorHeightFlag,
		},
	}

	for name, tc := range testcases {
		s.T().Run(name, func(t *testing.T) {
			// Setup
			clientCtx := val0.ClientCtx
			clientCtx.GRPCConcurrency = true
			clientCtx.Height = 0

			if tc.clientContextHeight != heightNotSetFlag {
				clientCtx = clientCtx.WithHeight(tc.clientContextHeight)
			}

			grpcContext := context.Background()
			if tc.grpcHeight != heightNotSetFlag {
				header := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", tc.grpcHeight))
				grpcContext = metadata.NewOutgoingContext(grpcContext, header)
			}

			// Test
			var header metadata.MD
			denom := fmt.Sprintf("%stoken", val0.Moniker)
			request := &banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom}
			response := &banktypes.QueryBalanceResponse{}
			err := clientCtx.Invoke(grpcContext, method, request, response, grpc.Header(&header))

			// Assert results
			if tc.expectedHeight == errorHeightFlag {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(
				sdk.NewCoin(denom, s.network.Config.AccountTokens),
				*response.GetBalance(),
			)
			blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
			s.Require().Equal([]string{fmt.Sprintf("%d", tc.expectedHeight)}, blockHeight)
		})
	}
}

func TestSelectHeight(t *testing.T) {
	testcases := map[string]testcase{
		"clientContextHeight 1; grpcHeight not set - clientContextHeight selected": {
			clientContextHeight: 1,
			grpcHeight:          heightNotSetFlag,
			expectedHeight:      1,
		},
		"clientContextHeight not set; grpcHeight is 2 - grpcHeight is chosen": {
			clientContextHeight: heightNotSetFlag,
			grpcHeight:          2,
			expectedHeight:      2,
		},
		"both not set - 0 returned": {
			clientContextHeight: heightNotSetFlag,
			grpcHeight:          heightNotSetFlag,
			expectedHeight:      0,
		},
		"clientContextHeight 3; grpcHeight is 0 - grpcHeight is chosen": {
			clientContextHeight: 3,
			grpcHeight:          0,
			expectedHeight:      0,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			clientCtx := client.Context{}
			clientCtx.GRPCConcurrency = true
			if tc.clientContextHeight != heightNotSetFlag {
				clientCtx = clientCtx.WithHeight(tc.clientContextHeight)
			}

			grpcContxt := context.Background()
			if tc.grpcHeight != heightNotSetFlag {
				header := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", tc.grpcHeight))
				grpcContxt = metadata.NewOutgoingContext(grpcContxt, header)
			}

			height, err := client.SelectHeight(clientCtx, grpcContxt)
			require.NoError(t, err)
			require.Equal(t, tc.expectedHeight, height)
		})
	}
}
