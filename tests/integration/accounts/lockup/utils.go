package lockup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/simapp"
	types "cosmossdk.io/x/accounts/defaults/lockup/v1"
	"cosmossdk.io/x/bank/testutil"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ownerAddr = secp256k1.GenPrivKey().PubKey().Address()
	accOwner  = sdk.AccAddress(ownerAddr)
)

type IntegrationTestSuite struct {
	suite.Suite

	app *simapp.SimApp
}

func NewIntegrationTestSuite() *IntegrationTestSuite {
	return &IntegrationTestSuite{}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	s.app = setupApp(s.T())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func setupApp(t *testing.T) *simapp.SimApp {
	t.Helper()
	app := simapp.Setup(t, false)
	return app
}

func (s *IntegrationTestSuite) executeTx(ctx sdk.Context, msg sdk.Msg, app *simapp.SimApp, accAddr, sender []byte) error {
	_, err := app.AccountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *IntegrationTestSuite) queryAcc(ctx sdk.Context, req sdk.Msg, app *simapp.SimApp, accAddr []byte) (transaction.Msg, error) {
	resp, err := app.AccountsKeeper.Query(ctx, accAddr, req)
	return resp, err
}

func (s *IntegrationTestSuite) fundAccount(app *simapp.SimApp, ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) {
	require.NoError(s.T(), testutil.FundAccount(ctx, app.BankKeeper, addr, amt))
}

func (s *IntegrationTestSuite) queryLockupAccInfo(ctx sdk.Context, app *simapp.SimApp, accAddr []byte) *types.QueryLockupAccountInfoResponse {
	req := &types.QueryLockupAccountInfoRequest{}
	resp, err := s.queryAcc(ctx, req, app, accAddr)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)

	lockupAccountInfoResponse, ok := resp.(*types.QueryLockupAccountInfoResponse)
	require.True(s.T(), ok)

	return lockupAccountInfoResponse
}

func (s *IntegrationTestSuite) queryUnbondingEntries(ctx sdk.Context, app *simapp.SimApp, accAddr []byte, valAddr string) *types.QueryUnbondingEntriesResponse {
	req := &types.QueryUnbondingEntriesRequest{
		ValidatorAddress: valAddr,
	}
	resp, err := s.queryAcc(ctx, req, app, accAddr)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)

	unbondingEntriesResponse, ok := resp.(*types.QueryUnbondingEntriesResponse)
	require.True(s.T(), ok)

	return unbondingEntriesResponse
}

func (s *IntegrationTestSuite) setupStakingParams(ctx sdk.Context, app *simapp.SimApp) {
	params, err := app.StakingKeeper.Params.Get(ctx)
	require.NoError(s.T(), err)

	// update unbonding time
	params.UnbondingTime = time.Duration(time.Second * 10)
	err = app.StakingKeeper.Params.Set(ctx, params)
	require.NoError(s.T(), err)
}
