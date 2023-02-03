//go:build norace
// +build norace

package client_test

import (
	"context"
	"fmt"
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

func (s *IntegrationTestSuite) TestGRPCQuery() {
	val0 := s.network.Validators[0]

	// gRPC query to test service should work
	testClient := testdata.NewQueryClient(val0.ClientCtx)
	testRes, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)

	// gRPC query to bank service should work
	denom := fmt.Sprintf("%stoken", val0.Moniker)
	bankClient := banktypes.NewQueryClient(val0.ClientCtx)
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
	val0.ClientCtx = val0.ClientCtx.WithHeight(1) // We set clientCtx to height 1
	bankClient = banktypes.NewQueryClient(val0.ClientCtx)
	bankRes, err = bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
		grpc.Header(&header),
	)
	blockHeight = header.Get(grpctypes.GRPCBlockHeightHeader)
	s.Require().Equal([]string{"1"}, blockHeight)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
