package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/counter"
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

	logger := log.NewNopLogger()
	cms := integration.CreateMultiStore(storetypes.NewKVStoreKeys(countertypes.StoreKey), logger)
	s.ctx = sdk.NewContext(cms, true, logger)
	cfg := moduletestutil.MakeTestEncodingConfig(counter.AppModuleBasic{})
	s.cdc = cfg.Codec

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, cfg.InterfaceRegistry)
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
	res, err := s.counterClient.GetCount(s.ctx, &countertypes.QueryGetCountRequest{}, grpc.Header(&header))
	s.Require().NoError(err)
	s.Require().Equal(0, res.TotalCount)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
