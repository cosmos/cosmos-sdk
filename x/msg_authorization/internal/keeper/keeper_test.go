package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type TestSuite struct {
	suite.Suite
	ctx           sdk.Context
	accountKeeper auth.AccountKeeper
	paramsKeeper  params.Keeper
	bankKeeper    bank.Keeper
	keeper        Keeper
	router        sdk.Router
}

func (s *TestSuite) SetupTest() {
	s.ctx, s.accountKeeper, s.paramsKeeper, s.bankKeeper, s.keeper, s.router = SetupTestInput()

}

func (s *TestSuite) TestKeeper(t *testing.T) {
	err := s.bankKeeper.SetCoins(s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	require.Nil(t, err)
	require.True(t, s.bankKeeper.GetCoins(s.ctx, granterAddr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("steak", 10000))))

	t.Log("Verify that no capability returns nil")
	capability, _ := s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(capability)
	//require.Nil(t, expiration)
	now := s.ctx.BlockHeader().Time
	require.NotNil(t, now)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.keeper.Grant(s.ctx, granterAddr, granteeAddr, types.SendCapability{SpendLimit: newCoins}, now.Add(-1*time.Hour))
	capability, expiration := s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	require.NotNil(t, capability)
	require.NotNil(t, expiration)
	//TODO add more cases
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
