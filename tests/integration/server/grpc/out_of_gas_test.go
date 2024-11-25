package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"

	storetypes "cosmossdk.io/store/types"
	minttypes "cosmossdk.io/x/mint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type IntegrationTestOutOfGasSuite struct {
	suite.Suite

	conn    *grpc.ClientConn
	address sdk.AccAddress
}

func (s *IntegrationTestOutOfGasSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(s.T())
	authority := authtypes.NewModuleAddress("gov")

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)
	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (uint64, error) {
		currentNum := accNum
		accNum++
		return currentNum, nil
	})

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		acctsModKeeper,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		accountKeeper,
		blockedAddresses,
		authority.String(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)

	integrationApp := integration.NewIntegrationApp(logger, keys, cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName: authModule,
			banktypes.ModuleName: bankModule,
		},
		baseapp.NewMsgServiceRouter(),
		baseapp.NewGRPCQueryRouter(),
		// baseapp.SetQueryGasLimit(10),
	)

	pubkeys := simtestutil.CreateTestPubKeys(2)
	s.address = sdk.AccAddress(pubkeys[0].Address())
	addr2 := sdk.AccAddress(pubkeys[1].Address())

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// mint some tokens
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	s.Require().NoError(bankKeeper.MintCoins(sdkCtx, minttypes.ModuleName, amount))
	s.Require().NoError(bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, minttypes.ModuleName, addr2, amount))

	// Register MsgServer and QueryServer
	banktypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), bankkeeper.NewMsgServerImpl(bankKeeper))
	banktypes.RegisterQueryServer(integrationApp.GRPCQueryRouter(), bankkeeper.NewQuerier(&bankKeeper))
	testdata.RegisterQueryServer(integrationApp.GRPCQueryRouter(), testdata.QueryImpl{})
	banktypes.RegisterInterfaces(encodingCfg.InterfaceRegistry)

	_, err := integrationApp.RunMsg(
		&banktypes.MsgSend{
			FromAddress: addr2.String(),
			ToAddress:   s.address.String(),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
		},
		integration.WithAutomaticFinalizeBlock(),
		integration.WithAutomaticCommit(),
	)
	s.Require().NoError(err)
	s.Require().Equal(integrationApp.LastBlockHeight(), int64(2))

	resp, err := bankKeeper.Balance(integrationApp.Context(), &banktypes.QueryBalanceRequest{Address: s.address.String(), Denom: "stake"})
	s.Require().NoError(err)
	s.Require().Equal(int64(50), resp.Balance.Amount.Int64())

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(encodingCfg.InterfaceRegistry).GRPCCodec()),
	)
	integrationApp.RegisterGRPCServer(grpcSrv)

	grpcCfg := srvconfig.DefaultConfig().GRPC
	go func() {
		err := servergrpc.StartGRPCServer(
			integrationApp.Context(),
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
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(encodingCfg.InterfaceRegistry).GRPCCodec())),
	)
	s.Require().NoError(err)
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

	s.Require().Equal(math.NewInt(50).Int64(), res.Balance.Amount.Int64())
	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

func TestIntegrationTestOutOfGasSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestOutOfGasSuite))
}
