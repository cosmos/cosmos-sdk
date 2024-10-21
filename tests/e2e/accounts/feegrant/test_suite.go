package feegrant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/simapp"
	"cosmossdk.io/x/bank/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type E2ETestSuite struct {
	suite.Suite

	app *simapp.SimApp
}

func NewE2ETestSuite() *E2ETestSuite {
	return &E2ETestSuite{}
}

func (s *E2ETestSuite) SetupSuite() {
	s.app = setupApp(s.T())
}

func (s *E2ETestSuite) TearDownSuite() {}

func setupApp(t *testing.T) *simapp.SimApp {
	t.Helper()
	app := simapp.Setup(t, false)
	return app
}

func (s *E2ETestSuite) ExecuteTX(ctx context.Context, msg transaction.Msg, accAddr, sender []byte) error {
	_, err := s.app.AccountsKeeper.Execute(ctx, accAddr, sender, msg, nil)
	return err
}

func (s *E2ETestSuite) queryAcc(ctx context.Context, req sdk.Msg, accAddr []byte) (transaction.Msg, error) {
	resp, err := s.app.AccountsKeeper.Query(ctx, accAddr, req)
	return resp, err
}

func (s *E2ETestSuite) fundAccount(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) {
	require.NoError(s.T(), testutil.FundAccount(ctx, s.app.BankKeeper, addr, amt))
}

func must[T any](r T, err error) T {
	if err != nil {
		panic(err)
	}
	return r
}
