package grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"

	_ "cosmossdk.io/x/accounts"                       // import as blank for app wiring
	_ "cosmossdk.io/x/bank"                           // import as blank for app wiring
	_ "cosmossdk.io/x/consensus"                      // import as blank for app wiring
	_ "cosmossdk.io/x/mint"                           // import as blank for app wiring
	_ "cosmossdk.io/x/staking"                        // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring

	minttypes "cosmossdk.io/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

type IntegrationTestOutOfGasSuite struct {
	suite.Suite

	conn    *grpc.ClientConn
	address sdk.AccAddress
	cdc     codec.Codec
	ctx     context.Context

	authKeeper authkeeper.AccountKeeper
	bankKeeper bankkeeper.BaseKeeper
}

func (s *IntegrationTestOutOfGasSuite) SetupSuite() {
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

	// startupCfg.BranchService = &integration.BranchService{}
	// startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}

	integrationApp, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...),
			depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&s.bankKeeper, &s.authKeeper, &s.cdc)
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
		grpc.ForceServerCodec(codec.NewProtoCodec(s.cdc.InterfaceRegistry()).GRPCCodec()),
	)
	// banktypes.RegisterQueryServer(grpcSrv, bankkeeper.NewQuerier(&s.bankKeeper))
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
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(s.cdc.InterfaceRegistry()).GRPCCodec())),
	)
	s.Require().NoError(err)
}

func (s *IntegrationTestOutOfGasSuite) registerQueryRouterService(router *integration.RouterService) {
	// register custom router service
	// queryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
	// 	req, ok := msg.(*banktypes.QueryBalanceRequest)
	// 	if !ok {
	// 		return nil, integration.ErrInvalidMsgType
	// 	}
	// 	qs := bankkeeper.NewQuerier(&s.bankKeeper)
	// 	resp, err := qs.Balance(ctx, req)
	// 	return resp, err
	// }

	// router.RegisterHandler(queryHandler, "cosmos.bank.v1beta1.Balance")
}

func (s *IntegrationTestOutOfGasSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.conn.Close()
}

func (s *IntegrationTestOutOfGasSuite) TestGRPCServer_TestService() {
	// gRPC query to test service should work
	testClient := testdata.NewQueryClient(s.conn)
	testRes, err := testClient.Echo(
		context.Background(),
		&testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)
}

func (s *IntegrationTestOutOfGasSuite) TestGRPCServer_BankBalance_OutOfGas() {
	// gRPC query to bank service should work
	bankClient := banktypes.NewQueryClient(s.conn)

	res, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: s.address.String(), Denom: "stake"},
	)

	fmt.Println("Res....", res, "Err.........", err)

	s.Require().Equal(math.NewInt(50).Int64(), res.Balance.Amount.Int64())
	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

func TestIntegrationTestOutOfGasSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestOutOfGasSuite))
}
