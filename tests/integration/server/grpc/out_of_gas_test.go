package grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type IntegrationTestOutOfGasSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	conn    *grpc.ClientConn
}

func (s *IntegrationTestOutOfGasSuite) SetupSuite() {
	var err error
	s.T().Log("setting up integration test suite")

	s.cfg, err = network.DefaultConfigWithAppConfigWithQueryGasLimit(configurator.NewAppConfig(
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.GenutilModule(),
		configurator.StakingModule(),
		configurator.ConsensusModule(),
		configurator.TxModule(),
	), 10)
	s.NoError(err)
	s.cfg.NumValidators = 1

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(2)
	s.Require().NoError(err)

	val0 := s.network.Validators[0]
	s.conn, err = grpc.NewClient(
		val0.GetAppConfig().GRPC.Address,
		grpc.WithInsecure(), //nolint:staticcheck // ignore SA1019, we don't need to use a secure connection for tests
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(s.cfg.InterfaceRegistry).GRPCCodec())),
	)
	s.Require().NoError(err)
}

func (s *IntegrationTestOutOfGasSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.conn.Close()
	s.network.Cleanup()
}

func (s *IntegrationTestOutOfGasSuite) TestGRPCServer_TestService() {
	// gRPC query to test service should work - simple queries should stay under gas limit
	testClient := testdata.NewQueryClient(s.conn)
	testRes, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)
}

func (s *IntegrationTestOutOfGasSuite) TestGRPCServer_BankBalance_OutOfGas() {
	val0 := s.network.Validators[0]

	// gRPC query to bank service should work
	denom := fmt.Sprintf("%stoken", val0.Moniker)
	bankClient := banktypes.NewQueryClient(s.conn)
	var header metadata.MD
	_, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: val0.Address.String(), Denom: denom},
		grpc.Header(&header), // Also fetch grpc header
	)

	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

func (s *IntegrationTestOutOfGasSuite) TestGRPCServer_AllBalances_OutOfGas() {
	val0 := s.network.Validators[0]

	// More complex query that requires more gas - querying all balances
	bankClient := banktypes.NewQueryClient(s.conn)
	_, err := bankClient.AllBalances(
		context.Background(),
		&banktypes.QueryAllBalancesRequest{Address: val0.Address.String()},
	)

	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

func (s *IntegrationTestOutOfGasSuite) TestGRPCServer_StakingQueries_OutOfGas() {
	// Test another module's queries to ensure the gas limit applies there too
	stakingClient := stakingtypes.NewQueryClient(s.conn)

	// This query should be complex enough to exceed our tiny gas limit
	_, err := stakingClient.Validators(
		context.Background(),
		&stakingtypes.QueryValidatorsRequest{},
	)

	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

func TestIntegrationTestOutOfGasSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestOutOfGasSuite))
}
