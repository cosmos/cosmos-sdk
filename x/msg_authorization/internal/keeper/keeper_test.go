package keeper

import (
	"testing"
	"time"

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

func (s *TestSuite) TestKeeper() {
	err := s.bankKeeper.SetCoins(s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	s.Require().Nil(err)
	s.Require().True(s.bankKeeper.GetCoins(s.ctx, granterAddr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("steak", 10000))))

	s.T().Log("Verify that no capability returns nil")
	capability, _ := s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(capability)
	//require.Nil(t, expiration)
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("Verify if expired capability is rejected")
	s.keeper.Grant(s.ctx, granterAddr, granteeAddr, types.SendCapability{SpendLimit: newCoins}, now.Add(-1*time.Hour))
	capability, _ = s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(capability)

	s.T().Log("Verify if capability is accepted")
	s.keeper.Grant(s.ctx, granteeAddr, granterAddr, types.SendCapability{SpendLimit: newCoins}, now.Add(time.Hour))
	capability, _ = s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(capability)
	s.Require().Equal(capability.MsgType(), bank.MsgSend{})

	s.T().Log("Verify fetching capability with wrong msg type fails")
	capability, _ = s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgMultiSend{})
	s.Require().Nil(capability)

	s.T().Log("Verify fetching capability with wrong grantee fails")
	capability, _ = s.keeper.GetCapability(s.ctx, recepientAddr, granterAddr, bank.MsgMultiSend{})
	s.Require().Nil(capability)

	s.T().Log("Verify revoke fails with wrong information")
	s.keeper.Revoke(s.ctx, recepientAddr, granterAddr, bank.MsgSend{})
	capability, _ = s.keeper.GetCapability(s.ctx, recepientAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(capability)

	s.T().Log("Verify revoke executes with correct information")
	s.keeper.Revoke(s.ctx, recepientAddr, granterAddr, bank.MsgSend{})
	capability, _ = s.keeper.GetCapability(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(capability)

}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
