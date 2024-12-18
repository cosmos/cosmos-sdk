package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	reflectionv1 "github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/codec"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	reflectionv2 "github.com/cosmos/cosmos-sdk/server/grpc/reflection/v2alpha1"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	_ "cosmossdk.io/x/accounts"                       // import as blank for app wiring
	_ "cosmossdk.io/x/bank"                           // import as blank for app wiring
	_ "cosmossdk.io/x/consensus"                      // import as blank for app wiring
	_ "cosmossdk.io/x/mint"                           // import as blank for app wiring
	_ "cosmossdk.io/x/staking"                        // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

type IntegrationTestSuite struct {
	suite.Suite

	conn    *grpc.ClientConn
	address sdk.AccAddress
	codec   codec.Codec
	ctx     context.Context

	authKeeper    authkeeper.AccountKeeper
	bankKeeper    bankkeeper.BaseKeeper
	stakingKeeper *stakingkeeper.Keeper
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
		configurator.MintModule(),
	}

	startupCfg := integration.DefaultStartUpConfig(s.T())

	// var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
	// 	return integration.NewRouterService()
	// }
	// queryRouterService := integration.NewRouterService()
	// s.registerQueryRouterService(queryRouterService)
	// serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	// startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}

	integrationApp, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...),
			depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&s.bankKeeper, &s.authKeeper, &s.codec, &s.stakingKeeper)
	s.Require().NoError(err)

	pubkeys := simtestutil.CreateTestPubKeys(2)
	s.address = sdk.AccAddress(pubkeys[0].Address())
	addr2 := sdk.AccAddress(pubkeys[1].Address())

	s.ctx = integrationApp.StateLatestContext(s.T())

	// mint some tokens
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	s.Require().NoError(s.bankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amount))
	s.Require().NoError(s.bankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr2, amount))

	bankMsgServer := bankkeeper.NewMsgServerImpl(s.bankKeeper)

	_, err = integrationApp.RunMsg(
		s.T(),
		s.ctx,
		func(ctx context.Context) (transaction.Msg, error) {
			resp, e := bankMsgServer.Send(ctx, &banktypes.MsgSend{
				FromAddress: addr2.String(),
				ToAddress:   s.address.String(),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
			})
			return resp, e
		},
		integration.WithAutomaticCommit(),
	)
	s.Require().NoError(err)
	// s.Require().Equal(integrationApp.LastBlockHeight(), int64(2))

	resp, err := s.bankKeeper.Balance(s.ctx, &banktypes.QueryBalanceRequest{Address: s.address.String(), Denom: "stake"})
	s.Require().NoError(err)
	s.Require().Equal(int64(50), resp.Balance.Amount.Int64())

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(s.codec.InterfaceRegistry()).GRPCCodec()),
	)
	banktypes.RegisterQueryServer(grpcSrv, bankkeeper.NewQuerier(&s.bankKeeper))
	stakingtypes.RegisterQueryServer(grpcSrv, stakingkeeper.NewQuerier(s.stakingKeeper))
	// integrationApp.RegisterGRPCServer(grpcSrv)

	grpcCfg := srvconfig.DefaultConfig().GRPC
	go func() {
		err := servergrpc.StartGRPCServer(
			s.ctx,
			integrationApp.Logger(),
			grpcCfg,
			grpcSrv,
		)
		s.Require().NoError(err)
		defer grpcSrv.GracefulStop()
	}()

	s.conn, err = grpc.NewClient(
		grpcCfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(s.codec.InterfaceRegistry()).GRPCCodec())),
	)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.conn.Close()
}

func (s *IntegrationTestSuite) TestGRPCServer_TestService() {
	// gRPC query to test service should work
	testClient := testdata.NewQueryClient(s.conn)
	testRes, err := testClient.Echo(
		context.Background(),
		&testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)
}

func (s *IntegrationTestSuite) TestGRPCServer_BankBalance_OutOfGas() {
	// gRPC query to bank service should work
	bankClient := banktypes.NewQueryClient(s.conn)

	_, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: s.address.String(), Denom: "stake"},
	)
	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

// Test and enforce that we upfront reject any connections to baseapp containing
// invalid initial x-cosmos-block-height that aren't positive  and in the range [0, max(int64)]
// See issue https://github.com/cosmos/cosmos-sdk/issues/7662.
func (s *IntegrationTestSuite) TestGRPCServerInvalidHeaderHeights() {
	t := s.T()

	// We should reject connections with invalid block heights off the bat.
	invalidHeightStrs := []struct {
		value   string
		wantErr string
	}{
		{"-1", "height < 0"},
		{"9223372036854775808", "value out of range"}, // > max(int64) by 1
		{"-10", "height < 0"},
		{"18446744073709551615", "value out of range"}, // max uint64, which is  > max(int64)
		{"-9223372036854775809", "value out of range"}, // Out of the range of for negative int64
	}
	for _, tt := range invalidHeightStrs {
		t.Run(tt.value, func(t *testing.T) {
			testClient := testdata.NewQueryClient(s.conn)
			ctx := metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, tt.value)
			testRes, err := testClient.Echo(ctx, &testdata.EchoRequest{Message: "hello"})
			require.Error(t, err)
			require.Nil(t, testRes)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func (s *IntegrationTestSuite) TestGRPCServer_Reflection() {
	// Test server reflection
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// NOTE(fdymylja): we use grpcreflect because it solves imports too
	// so that we can always assert that given a reflection server it is
	// possible to fully query all the methods, without having any context
	// on the proto registry
	rc := grpcreflect.NewClientAuto(ctx, s.conn)

	services, err := rc.ListServices()
	s.Require().NoError(err)
	s.Require().Greater(len(services), 0)

	for _, svc := range services {
		file, err := rc.FileContainingSymbol(svc)
		s.Require().NoError(err)
		sd := file.FindSymbol(svc)
		s.Require().NotNil(sd)
	}
}

func (s *IntegrationTestSuite) TestGRPCServer_InterfaceReflection() {
	// s.T().Skip() // TODO: fix this test at https://github.com/cosmos/cosmos-sdk/issues/22825

	// this tests the application reflection capabilities and compatibility between v1 and v2
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientV2 := reflectionv2.NewReflectionServiceClient(s.conn)
	clientV1 := reflectionv1.NewReflectionServiceClient(s.conn)
	codecDesc, err := clientV2.GetCodecDescriptor(ctx, nil)
	s.Require().NoError(err)

	interfaces, err := clientV1.ListAllInterfaces(ctx, nil)
	s.Require().NoError(err)
	s.Require().Equal(len(codecDesc.Codec.Interfaces), len(interfaces.InterfaceNames))
	s.Require().Equal(len(s.codec.InterfaceRegistry().ListAllInterfaces()), len(codecDesc.Codec.Interfaces))

	for _, iface := range interfaces.InterfaceNames {
		impls, err := clientV1.ListImplementations(ctx, &reflectionv1.ListImplementationsRequest{InterfaceName: iface})
		s.Require().NoError(err)

		s.Require().ElementsMatch(impls.ImplementationMessageNames, s.codec.InterfaceRegistry().ListImplementations(iface))
	}
}

// TestGRPCUnpacker - tests the grpc endpoint for Validator and using the interface registry unpack and extract the
// ConsAddr. (ref: https://github.com/cosmos/cosmos-sdk/issues/8045)
func (s *IntegrationTestSuite) TestGRPCUnpacker() {
	queryClient := stakingtypes.NewQueryClient(s.conn)
	validators, err := queryClient.Validators(s.ctx, &stakingtypes.QueryValidatorsRequest{})
	require.NoError(s.T(), err)

	// if len(validators.Validators) == 0 {
	// 	s.T().Skip("no validators found")
	// }

	validator, err := queryClient.Validator(
		s.ctx,
		&stakingtypes.QueryValidatorRequest{ValidatorAddr: validators.Validators[0].OperatorAddress},
	)
	require.NoError(s.T(), err)

	// no unpacked interfaces yet, so ConsAddr will be nil
	nilAddr, err := validator.Validator.GetConsAddr()
	require.Error(s.T(), err)
	require.Nil(s.T(), nilAddr)

	// unpack the interfaces and now ConsAddr is not nil
	err = validator.Validator.UnpackInterfaces(s.codec.InterfaceRegistry())
	require.NoError(s.T(), err)
	addr, err := validator.Validator.GetConsAddr()
	require.NotNil(s.T(), addr)
	require.NoError(s.T(), err)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
