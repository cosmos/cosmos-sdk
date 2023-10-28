package client_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	counterkeeper "github.com/cosmos/cosmos-sdk/x/counter/keeper"
	countertypes "github.com/cosmos/cosmos-sdk/x/counter/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	cdc           codec.Codec
	counterClient countertypes.QueryClient
	testClient    testdata.QueryClient
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	var (
		interfaceRegistry codectypes.InterfaceRegistry
		cdc               codec.Codec
	)

	logger := log.NewNopLogger()

	keys := storetypes.NewKVStoreKeys(countertypes.StoreKey)
	cms := integration.CreateMultiStore(keys, logger)

	s.ctx = sdk.NewContext(cms, true, logger)
	s.cdc = cdc
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, interfaceRegistry)
	testdata.RegisterQueryServer(queryHelper, testdata.QueryImpl{})
	countertypes.RegisterQueryServer(queryHelper, counterkeeper.Keeper{})
	s.counterClient = countertypes.NewQueryClient(queryHelper)
	s.testClient = testdata.NewQueryClient(queryHelper)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestGRPCQuery() {
	// gRPC query to test service should work
	testRes, err := s.testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)

	var header metadata.MD
	res, err := s.counterClient.GetCount(s.ctx, &countertypes.QueryCountRequest{}, grpc.Header(&header))
	s.Require().NoError(err)
	s.Require().Equal(0, res.TotalCount)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
