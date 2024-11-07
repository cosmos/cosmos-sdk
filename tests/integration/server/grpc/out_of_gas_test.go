package grpc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
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

	grpcCtx context.Context
	conn    *grpc.ClientConn
	address sdk.AccAddress
}

func (s *IntegrationTestOutOfGasSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(s.T())
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

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

	s.Require().NoError(bankKeeper.SetParams(newCtx, banktypes.DefaultParams()))

	authModule := auth.NewAppModule(cdc, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc,
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

	grpcCfg := srvconfig.DefaultConfig().GRPC

	msgSend := banktypes.MsgSend{
		FromAddress: addr2.String(),
		ToAddress:   s.address.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
	}

	_, err := integrationApp.RunMsg(
		&msgSend,
		integration.WithAutomaticFinalizeBlock(),
		integration.WithAutomaticCommit(),
	)
	s.Require().NoError(err)

	grpcSrv := grpc.NewServer(grpc.ForceServerCodec(codec.NewProtoCodec(encodingCfg.InterfaceRegistry).GRPCCodec()))
	integrationApp.RegisterGRPCServer(grpcSrv)

	go func() {
		s.Require().NoError(servergrpc.StartGRPCServer(context.Background(), integrationApp.Logger(), grpcCfg, grpcSrv))
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

	fmt.Println("Res....", res)

	s.Require().ErrorContains(err, sdkerrors.ErrOutOfGas.Error())
}

func TestIntegrationTestOutOfGasSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestOutOfGasSuite))
}
