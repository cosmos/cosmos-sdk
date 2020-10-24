// +build norace

package grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.network = network.New(s.T(), network.DefaultConfig())
	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(2)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGRPCServer() {
	val0 := s.network.Validators[0]
	conn, err := grpc.Dial(
		val0.AppConfig.GRPC.Address,
		grpc.WithInsecure(), // Or else we get "no transport security set"
	)
	s.Require().NoError(err)
	defer conn.Close()

	// gRPC query to test service should work
	testClient := testdata.NewQueryClient(conn)
	testRes, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)

	// gRPC query to bank service should work
	denom := fmt.Sprintf("%stoken", val0.Moniker)
	bankClient := banktypes.NewQueryClient(conn)
	var header metadata.MD
	bankRes, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
		grpc.Header(&header), // Also fetch grpc header
	)
	s.Require().NoError(err)
	s.Require().Equal(
		sdk.NewCoin(denom, s.network.Config.AccountTokens),
		*bankRes.GetBalance(),
	)
	blockHeight := header.Get(grpctypes.GRPCBlockHeightHeader)
	s.Require().NotEmpty(blockHeight[0]) // Should contain the block height

	// Request metadata should work
	bankRes, err = bankClient.Balance(
		metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, "1"), // Add metadata to request
		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
		grpc.Header(&header),
	)
	blockHeight = header.Get(grpctypes.GRPCBlockHeightHeader)
	s.Require().Equal([]string{"1"}, blockHeight)

	// Test server reflection
	reflectClient := rpb.NewServerReflectionClient(conn)
	stream, err := reflectClient.ServerReflectionInfo(context.Background(), grpc.WaitForReady(true))
	s.Require().NoError(err)
	s.Require().NoError(stream.Send(&rpb.ServerReflectionRequest{
		MessageRequest: &rpb.ServerReflectionRequest_ListServices{},
	}))
	res, err := stream.Recv()
	s.Require().NoError(err)
	services := res.GetListServicesResponse().Service
	servicesMap := make(map[string]bool)
	for _, s := range services {
		servicesMap[s.Name] = true
	}
	// Make sure the following services are present
	s.Require().True(servicesMap["cosmos.bank.v1beta1.Query"])
}

// Test and enforce that we upfront reject any connections to baseapp containing
// invalid initial x-cosmos-block-height that aren't positive  and in the range [0, max(int64)]
// See issue https://github.com/cosmos/cosmos-sdk/issues/7662.
func (s *IntegrationTestSuite) TestGRPCServerInvalidHeaderHeights() {
	t := s.T()
	val0 := s.network.Validators[0]

	// We should reject connections with invalid block heights off the bat.
	invalidHeightStrs := []string{
		"-1",
		"9223372036854775808", // > max(int64) by 1
		"-10",
		"18446744073709551615", // max uint64, which is  > max(int64)
		"-9223372036854775809", // Out of the range of for negative int64
	}
	for _, invalidHeightStr := range invalidHeightStrs {
		t.Run(invalidHeightStr, func(t *testing.T) {
			conn, err := grpc.Dial(
				val0.AppConfig.GRPC.Address,
				grpc.WithInsecure(), // Or else we get "no transport security set"
			)
			defer conn.Close()

			testClient := testdata.NewQueryClient(conn)
			ctx := metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, invalidHeightStr)
			testRes, err := testClient.Echo(ctx, &testdata.EchoRequest{Message: "hello"})
			require.Error(t, err)
			require.Nil(t, testRes)
			require.Contains(t, err.Error(), "value out of range")
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
