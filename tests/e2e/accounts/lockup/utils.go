package lockup

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"cosmossdk.io/x/bank/testutil"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ownerAddr = secp256k1.GenPrivKey().PubKey().Address()
	accOwner  = sdk.AccAddress(ownerAddr)
)

type E2ETestSuite struct {
	suite.Suite

	app *simapp.SimApp
}

func NewE2ETestSuite() *E2ETestSuite {
	return &E2ETestSuite{}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")
	s.app = setupApp(s.T())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
}

func setupApp(t *testing.T) *simapp.SimApp {
	t.Helper()
	app := simapp.Setup(t, false)
	return app
}

func (s *E2ETestSuite) executeTx(ctx sdk.Context, msg sdk.Msg, app *simapp.SimApp, accAddr, sender []byte) error {
	_, err := app.AccountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *E2ETestSuite) fundAccount(app *simapp.SimApp, ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) {
	require.NoError(s.T(), testutil.FundAccount(ctx, app.BankKeeper, addr, amt))
}
