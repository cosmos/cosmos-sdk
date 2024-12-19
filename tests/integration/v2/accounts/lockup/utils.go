package lockup

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/x/accounts"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	lockupdepinject "cosmossdk.io/x/accounts/defaults/lockup/depinject"
	types "cosmossdk.io/x/accounts/defaults/lockup/v1"
	_ "cosmossdk.io/x/bank" // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/distribution" // import as blank for app wiring
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	_ "cosmossdk.io/x/staking" // import as blank for app wiring
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

var (
	ownerAddr = secp256k1.GenPrivKey().PubKey().Address()
	accOwner  = sdk.AccAddress(ownerAddr)
)

type IntegrationTestSuite struct {
	suite.Suite

	app *integration.App
	ctx context.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.BaseKeeper
	stakingKeeper  *stakingkeeper.Keeper
	distrKeeper    distrkeeper.Keeper
}

func NewIntegrationTestSuite() *IntegrationTestSuite {
	return &IntegrationTestSuite{}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.VestingModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
		configurator.DistributionModule(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(s.T())

	msgRouterService := integration.NewRouterService()
	s.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := integration.NewRouterService()
	s.registerQueryRouterService(queryRouterService)

	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}
	startupCfg.GasService = &integration.GasService{}

	s.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Provide(
			// inject desired account types:
			basedepinject.ProvideAccount,

			// provide base account options
			basedepinject.ProvideSecp256K1PubKey,

			// inject desired account types:
			lockupdepinject.ProvideAllLockupAccounts,
		), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&s.bankKeeper, &s.accountsKeeper, &s.authKeeper, &s.stakingKeeper, &s.distrKeeper)
	require.NoError(s.T(), err)

	s.ctx = s.app.StateLatestContext(s.T())
}

func (s *IntegrationTestSuite) registerMsgRouterService(router *integration.RouterService) {
	// register custom router service
	bankSendHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*banktypes.MsgSend)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := bankkeeper.NewMsgServerImpl(s.bankKeeper)
		resp, err := msgServer.Send(ctx, msg)
		return resp, err
	}

	stakingDelegateHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*stakingtypes.MsgDelegate)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := stakingkeeper.NewMsgServerImpl(s.stakingKeeper)
		resp, err := msgServer.Delegate(ctx, msg)
		return resp, err
	}

	stakingUndelegateHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*stakingtypes.MsgUndelegate)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := stakingkeeper.NewMsgServerImpl(s.stakingKeeper)
		resp, err := msgServer.Undelegate(ctx, msg)
		return resp, err
	}

	distrWithdrawRewardHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*distrtypes.MsgWithdrawDelegatorReward)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := distrkeeper.NewMsgServerImpl(s.distrKeeper)
		resp, err := msgServer.WithdrawDelegatorReward(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
	router.RegisterHandler(stakingDelegateHandler, "cosmos.staking.v1beta1.MsgDelegate")
	router.RegisterHandler(stakingUndelegateHandler, "cosmos.staking.v1beta1.MsgUndelegate")
	router.RegisterHandler(distrWithdrawRewardHandler, "cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward")
}

func (s *IntegrationTestSuite) registerQueryRouterService(router *integration.RouterService) {
	// register custom router service
	stakingParamsQueryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*stakingtypes.QueryParamsRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := stakingkeeper.NewQuerier(s.stakingKeeper)
		resp, err := qs.Params(ctx, req)
		return resp, err
	}

	stakingUnbondingQueryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*stakingtypes.QueryUnbondingDelegationRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := stakingkeeper.NewQuerier(s.stakingKeeper)
		resp, err := qs.UnbondingDelegation(ctx, req)
		return resp, err
	}

	bankBalanceQueryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*banktypes.QueryBalanceRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := bankkeeper.NewQuerier(&s.bankKeeper)
		resp, err := qs.Balance(ctx, req)
		return resp, err
	}

	router.RegisterHandler(stakingParamsQueryHandler, "cosmos.staking.v1beta1.QueryParamsRequest")
	router.RegisterHandler(stakingUnbondingQueryHandler, "cosmos.staking.v1beta1.QueryUnbondingDelegationRequest")
	router.RegisterHandler(bankBalanceQueryHandler, "cosmos.bank.v1beta1.QueryBalanceRequest")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) executeTx(ctx context.Context, msg sdk.Msg, ak accounts.Keeper, accAddr, sender []byte) error {
	_, err := ak.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *IntegrationTestSuite) queryAcc(ctx context.Context, req sdk.Msg, ak accounts.Keeper, accAddr []byte) (transaction.Msg, error) {
	resp, err := ak.Query(ctx, accAddr, req)
	return resp, err
}

func (s *IntegrationTestSuite) fundAccount(bk bankkeeper.Keeper, ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) {
	require.NoError(s.T(), testutil.FundAccount(ctx, bk, addr, amt))
}

func (s *IntegrationTestSuite) queryLockupAccInfo(ctx context.Context, ak accounts.Keeper, accAddr []byte) *types.QueryLockupAccountInfoResponse {
	req := &types.QueryLockupAccountInfoRequest{}
	resp, err := s.queryAcc(ctx, req, ak, accAddr)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)

	lockupAccountInfoResponse, ok := resp.(*types.QueryLockupAccountInfoResponse)
	require.True(s.T(), ok)

	return lockupAccountInfoResponse
}

func (s *IntegrationTestSuite) queryUnbondingEntries(ctx context.Context, ak accounts.Keeper, accAddr []byte, valAddr string) *types.QueryUnbondingEntriesResponse {
	req := &types.QueryUnbondingEntriesRequest{
		ValidatorAddress: valAddr,
	}
	resp, err := s.queryAcc(ctx, req, ak, accAddr)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)

	unbondingEntriesResponse, ok := resp.(*types.QueryUnbondingEntriesResponse)
	require.True(s.T(), ok)

	return unbondingEntriesResponse
}

func (s *IntegrationTestSuite) setupStakingParams(ctx context.Context, sk *stakingkeeper.Keeper) {
	params, err := sk.Params.Get(ctx)
	require.NoError(s.T(), err)

	// update unbonding time
	params.UnbondingTime = time.Second * 10
	err = sk.Params.Set(ctx, params)
	require.NoError(s.T(), err)
}
