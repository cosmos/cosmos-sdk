package keeper

import (
	"fmt"
	"github.com/kr/pretty"
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

	s.T().Log("verify that no authorization returns nil")
	authorization, _ := s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(authorization)
	//require.Nil(t, expiration)
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin("steak", 100))
	s.T().Log("verify if expired authorization is rejected")
	x := types.SendAuthorization{SpendLimit: newCoins}
	s.keeper.Grant(s.ctx, granterAddr, granteeAddr, x, now.Add(-1*time.Hour))
	authorization, _ = s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(authorization)

	s.T().Log("verify if authorization is accepted")
	x := types.SendAuthorization{SpendLimit: newCoins}
	s.keeper.Grant(s.ctx, granteeAddr, granterAddr, x, now.Add(time.Hour))
	authorization, _ = s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(authorization)
	s.Require().Equal(authorization.MsgType(), bank.MsgSend{})

	s.T().Log("verify fetching authorization with wrong msg type fails")
	authorization, _ = s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgMultiSend{})
	s.Require().Nil(authorization)

	s.T().Log("verify fetching authorization with wrong grantee fails")
	authorization, _ = s.keeper.GetAuthorization(s.ctx, recepientAddr, granterAddr, bank.MsgMultiSend{})
	s.Require().Nil(authorization)

	s.T().Log("")

	s.T().Log("verify revoke fails with wrong information")
	s.keeper.Revoke(s.ctx, recepientAddr, granterAddr, bank.MsgSend{})
	authorization, _ = s.keeper.GetAuthorization(s.ctx, recepientAddr, granterAddr, bank.MsgSend{})
	s.Require().Nil(authorization)

	s.T().Log("verify revoke executes with correct information")
	s.keeper.Revoke(s.ctx, recepientAddr, granterAddr, bank.MsgSend{})
	authorization, _ = s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(authorization)

}

func (s *TestSuite) TestKeeperFees() {
	err := s.bankKeeper.SetCoins(s.ctx, granterAddr, sdk.NewCoins(sdk.NewInt64Coin("steak", 10000)))
	s.Require().Nil(err)
	s.Require().True(s.bankKeeper.GetCoins(s.ctx, granterAddr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("steak", 10000))))

	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)

	smallCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 20))
	someCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 123))
	//lotCoin := sdk.NewCoins(sdk.NewInt64Coin("steak", 4567))

	msgs := types.MsgExecDelegated{
		Grantee: granteeAddr,
		Msgs: []sdk.Msg{
			bank.MsgSend{
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
				FromAddress: granterAddr,
				ToAddress:   recepientAddr,
			},
		},
	}

	s.T().Log("verify dispatch fails with invalid authorization")
	result, error := s.keeper.DispatchActions(s.ctx, granteeAddr, msgs.Msgs)
	s.T().Log(error.Error())
	s.Require().Nil(result)
	s.Require().NotNil(error)

	s.T().Log("verify dispatch executes with correct information")
	// grant authorization
	s.keeper.Grant(s.ctx, granteeAddr, granterAddr, types.SendAuthorization{SpendLimit: smallCoin}, now)
	authorization, _ := s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(authorization)
	s.Require().Equal(authorization.MsgType(), bank.MsgSend{})
	result, error = s.keeper.DispatchActions(s.ctx, granteeAddr, msgs.Msgs)
	s.Require().NotNil(result)
	s.Require().Nil(error)
	fmt.Printf("%# v", pretty.Formatter(result))

	authorization, _ = s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(authorization)

	s.T().Log("verify dispatch fails with overlimit")
	// grant authorization

	msgs = types.MsgExecDelegated{
		Grantee: granteeAddr,
		Msgs: []sdk.Msg{
			bank.MsgSend{
				Amount:      someCoin,
				FromAddress: granterAddr,
				ToAddress:   recepientAddr,
			},
		},
	}

	result, error = s.keeper.DispatchActions(s.ctx, granteeAddr, msgs.Msgs)
	s.Require().Nil(result)
	s.Require().NotNil(error)
	fmt.Printf("%# v", pretty.Formatter(error))

	authorization, _ = s.keeper.GetAuthorization(s.ctx, granteeAddr, granterAddr, bank.MsgSend{})
	s.Require().NotNil(authorization)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
