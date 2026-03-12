package client_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx                   sdk.Context
	cdc                   codec.Codec
	genesisAccount        *authtypes.BaseAccount
	bankClient            types.QueryClient
	testClient            testdata.QueryClient
	genesisAccountBalance int64
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	var (
		interfaceRegistry codectypes.InterfaceRegistry
		bankKeeper        bankkeeper.BaseKeeper
		appBuilder        *runtime.AppBuilder
		cdc               codec.Codec
	)

	// TODO duplicated from testutils/sims/app_helpers.go
	// need more composable startup options for simapp, this test needed a handle to the closed over genesis account
	// to query balances
	err := depinject.Inject(
		depinject.Configs(
			testutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&interfaceRegistry, &bankKeeper, &appBuilder, &cdc)
	s.NoError(err)

	app := appBuilder.Build(dbm.NewMemDB(), nil)
	err = app.Load(true)
	s.NoError(err)

	valSet, err := sims.CreateRandomValidatorSet()
	s.NoError(err)

	// generate genesis account
	s.genesisAccountBalance = 100000000000000
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := types.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(s.genesisAccountBalance))),
	}

	genesisState, err := sims.GenesisStateWithValSet(cdc, app.DefaultGenesis(), valSet, []authtypes.GenesisAccount{acc}, balance)
	s.NoError(err)

	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	s.NoError(err)

	// init chain will set the validator set and initialize the genesis accounts
	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	s.NoError(err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})
	s.NoError(err)

	// end of app init

	s.ctx = app.NewContext(false)
	s.cdc = cdc
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, interfaceRegistry)
	types.RegisterQueryServer(queryHelper, bankKeeper)
	testdata.RegisterQueryServer(queryHelper, testdata.QueryImpl{})
	s.bankClient = types.NewQueryClient(queryHelper)
	s.testClient = testdata.NewQueryClient(queryHelper)
	s.genesisAccount = acc
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestGRPCQuery() {
	denom := sdk.DefaultBondDenom

	// gRPC query to test service should work
	testRes, err := s.testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)

	// gRPC query to bank service should work
	var header metadata.MD
	res, err := s.bankClient.Balance(
		context.Background(),
		&types.QueryBalanceRequest{Address: s.genesisAccount.GetAddress().String(), Denom: denom},
		grpc.Header(&header), // Also fetch grpc header
	)
	s.Require().NoError(err)
	bal := res.GetBalance()
	s.Equal(sdk.NewCoin(denom, math.NewInt(s.genesisAccountBalance)), *bal)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestGetGRPCConnWithContext() {
	defaultConn, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	s.Require().NoError(err)
	defer defaultConn.Close()

	historicalConn, err := grpc.NewClient("localhost:9091",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	s.Require().NoError(err)
	defer historicalConn.Close()

	historicalConns := config.HistoricalGRPCConnections{
		config.BlockRange{100, 500}: historicalConn,
	}
	provider := client.NewGRPCConnProvider(defaultConn, historicalConns)
	testCases := []struct {
		name         string
		height       int64
		setupCtx     func() client.Context
		expectedConn *grpc.ClientConn
	}{
		{
			name:   "context with GRPCConnProvider and historical height",
			height: 300,
			setupCtx: func() client.Context {
				return client.Context{}.
					WithCodec(s.cdc).
					WithGRPCClient(defaultConn).
					WithGRPCConnProvider(provider).
					WithHeight(300)
			},
			expectedConn: historicalConn,
		},
		{
			name:   "context with GRPCConnProvider and latest height",
			height: 0,
			setupCtx: func() client.Context {
				return client.Context{}.
					WithCodec(s.cdc).
					WithGRPCClient(defaultConn).
					WithGRPCConnProvider(provider).
					WithHeight(0)
			},
			expectedConn: defaultConn,
		},
		{
			name:   "context without GRPCConnProvider",
			height: 300,
			setupCtx: func() client.Context {
				return client.Context{}.
					WithCodec(s.cdc).
					WithGRPCClient(defaultConn).
					WithHeight(300)
			},
			expectedConn: defaultConn,
		},
		{
			name:   "context with nil historical connections map",
			height: 100,
			setupCtx: func() client.Context {
				nilProvider := client.NewGRPCConnProvider(defaultConn, nil)
				return client.Context{}.
					WithCodec(s.cdc).
					WithGRPCClient(defaultConn).
					WithGRPCConnProvider(nilProvider).
					WithHeight(100)
			},
			expectedConn: defaultConn,
		},
		{
			name:   "context with empty historical connections map",
			height: 100,
			setupCtx: func() client.Context {
				emptyProvider := client.NewGRPCConnProvider(defaultConn, config.HistoricalGRPCConnections{})
				return client.Context{}.
					WithCodec(s.cdc).
					WithGRPCClient(defaultConn).
					WithGRPCConnProvider(emptyProvider).
					WithHeight(100)
			},
			expectedConn: defaultConn,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := tc.setupCtx()
			var actualConn *grpc.ClientConn
			if ctx.GRPCConnProvider != nil {
				actualConn = ctx.GRPCConnProvider.GetGRPCConn(ctx.Height)
			} else {
				actualConn = ctx.GRPCClient
			}
			s.Require().Equal(tc.expectedConn, actualConn)
		})
	}
}

func TestGetHeightFromMetadata(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedHeight int64
	}{
		{
			name: "valid height in metadata",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "12345")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 12345,
		},
		{
			name: "zero height in metadata",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "0")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 0,
		},
		{
			name: "negative height returns zero",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "-100")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 0,
		},
		{
			name:           "no metadata returns zero",
			setupContext:   context.Background,
			expectedHeight: 0,
		},
		{
			name: "empty height header returns zero",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 0,
		},
		{
			name: "invalid height string returns zero",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "not-a-number")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 0,
		},
		{
			name: "multiple height values uses first",
			setupContext: func() context.Context {
				md := metadata.Pairs(
					grpctypes.GRPCBlockHeightHeader, "100",
					grpctypes.GRPCBlockHeightHeader, "200",
				)
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 100,
		},
		{
			name: "very large height",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "9223372036854775807") // max int64
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 9223372036854775807,
		},
		{
			name: "height exceeding int64 returns zero",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "9223372036854775808") // max int64 + 1
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			height := client.GetHeightFromMetadata(ctx)
			require.Equal(t, tt.expectedHeight, height)
		})
	}
}

func TestGetHeightFromMetadataStrict(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedHeight int64
		expectError    bool
	}{
		{
			name: "valid height",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "123")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectedHeight: 123,
		},
		{
			name:         "no metadata",
			setupContext: context.Background,
		},
		{
			name: "negative height errors",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "-10")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectError: true,
		},
		{
			name: "invalid height errors",
			setupContext: func() context.Context {
				md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, "foo")
				return metadata.NewOutgoingContext(context.Background(), md)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			height, err := client.GetHeightFromMetadataStrict(ctx)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedHeight, height)
		})
	}
}
