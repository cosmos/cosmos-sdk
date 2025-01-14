package multisig

import (
	"context"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/testing/msgrouter"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/x/accounts"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	multisigdepinject "cosmossdk.io/x/accounts/defaults/multisig/depinject"
	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	_ "cosmossdk.io/x/bank" // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/distribution" // import as blank for app wiring
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	_ "cosmossdk.io/x/staking" // import as blank for app wiring
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

type IntegrationTestSuite struct {
	suite.Suite

	app *integration.App

	members     []sdk.AccAddress
	membersAddr []string

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

	msgRouterService := msgrouter.NewRouterService()
	s.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := msgrouter.NewRouterService()
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
			multisigdepinject.ProvideAccount,
		), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&s.bankKeeper, &s.accountsKeeper, &s.authKeeper, &s.stakingKeeper, &s.distrKeeper)
	s.NoError(err)

	s.members = []sdk.AccAddress{}
	for i := 0; i < 10; i++ {
		addr := secp256k1.GenPrivKey().PubKey().Address()
		addrStr, err := s.authKeeper.AddressCodec().BytesToString(addr)
		s.NoError(err)
		s.membersAddr = append(s.membersAddr, addrStr)
		s.members = append(s.members, sdk.AccAddress(addr))
	}
}

func (s *IntegrationTestSuite) registerMsgRouterService(router *msgrouter.RouterService) {
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

	// register custom router service
	accountsExeccHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*accountsv1.MsgExecute)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := accounts.NewMsgServer(s.accountsKeeper)
		resp, err := msgServer.Execute(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
	router.RegisterHandler(accountsExeccHandler, "cosmos.accounts.v1.MsgExecute")
}

func (s *IntegrationTestSuite) registerQueryRouterService(router *msgrouter.RouterService) {
	// register custom router service
	bankBalanceQueryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*banktypes.QueryBalanceRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := bankkeeper.NewQuerier(&s.bankKeeper)
		resp, err := qs.Balance(ctx, req)
		return resp, err
	}

	router.RegisterHandler(bankBalanceQueryHandler, "cosmos.bank.v1beta1.QueryBalanceRequest")
}

func (s *IntegrationTestSuite) TearDownSuite() {}

func (s *IntegrationTestSuite) executeTx(ctx context.Context, msg sdk.Msg, accAddr, sender []byte) error {
	_, err := s.accountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *IntegrationTestSuite) queryAcc(ctx context.Context, req sdk.Msg, accAddr []byte) (transaction.Msg, error) {
	resp, err := s.accountsKeeper.Query(ctx, accAddr, req)
	return resp, err
}

func (s *IntegrationTestSuite) fundAccount(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) {
	s.NoError(testutil.FundAccount(ctx, s.bankKeeper, addr, amt))
}

// initAccount initializes a multisig account with the given members and powers
// and returns the account address
func (s *IntegrationTestSuite) initAccount(ctx context.Context, sender []byte, membersPowers map[string]uint64) ([]byte, string) {
	s.fundAccount(ctx, sender, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000000))})

	members := []*v1.Member{}
	for addrStr, power := range membersPowers {
		members = append(members, &v1.Member{Address: addrStr, Weight: power})
	}

	_, accountAddr, err := s.accountsKeeper.Init(ctx, "multisig", sender,
		&v1.MsgInit{
			Members: members,
			Config: &v1.Config{
				Threshold:      100,
				Quorum:         100,
				VotingPeriod:   120,
				Revote:         false,
				EarlyExecution: true,
			},
		}, sdk.Coins{sdk.NewCoin("stake", math.NewInt(1000))}, nil)
	s.NoError(err)

	accountAddrStr, err := s.authKeeper.AddressCodec().BytesToString(accountAddr)
	s.NoError(err)

	return accountAddr, accountAddrStr
}

// createProposal
func (s *IntegrationTestSuite) createProposal(ctx context.Context, accAddr, sender []byte, msgs ...*codectypes.Any) {
	propReq := &v1.MsgCreateProposal{
		Proposal: &v1.Proposal{
			Title:    "test",
			Summary:  "test",
			Messages: msgs,
		},
	}
	err := s.executeTx(ctx, propReq, accAddr, sender)
	s.NoError(err)
}

func (s *IntegrationTestSuite) executeProposal(ctx context.Context, accAddr, sender []byte, proposalID uint64) error {
	execReq := &v1.MsgExecuteProposal{
		ProposalId: proposalID,
	}
	return s.executeTx(ctx, execReq, accAddr, sender)
}
