package client_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	},
	)
	s.NoError(err)

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Hash:               app.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
	})
	s.NoError(err)

	// end of app init

	s.ctx = app.BaseApp.NewContext(false)
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
