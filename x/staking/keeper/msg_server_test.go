package keeper_test

import (
	"testing"
	"time"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestCancelUnbondingDelegation() {
	msgServer := keeper.NewMsgServerImpl(suite.stakingKeeper)
	bondDenom := suite.stakingKeeper.BondDenom(suite.ctx)

	// set the not bonded pool module account
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)
	startTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 5)

	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), startTokens))))
	suite.accountKeeper.SetModuleAccount(suite.ctx, notBondedPool)

	moduleBalance := suite.bankKeeper.GetBalance(suite.ctx, notBondedPool.GetAddress(), suite.stakingKeeper.BondDenom(suite.ctx))
	suite.Require().Equal(sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	delAddrs := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, sdk.NewInt(10000))
	validators := suite.stakingKeeper.GetValidators(suite.ctx, 10)
	suite.Require().Equal(len(validators), 1)

	validatorAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	suite.Require().NoError(err)
	delegatorAddr := delAddrs[0]

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(suite.stakingKeeper.BondDenom(suite.ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		suite.ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
	)

	// set and retrieve a record
	suite.stakingKeeper.SetUnbondingDelegation(suite.ctx, ubd)
	resUnbond, found := suite.stakingKeeper.GetUnbondingDelegation(suite.ctx, delegatorAddr, validatorAddr)
	suite.Require().True(found)
	suite.Require().Equal(ubd, resUnbond)

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCancelUnbondingDelegation
	}{
		{
			Name:      "invalid height",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid coin",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin("dump_coin", sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "validator not exists",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid amount",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Add(sdk.NewInt64Coin(bondDenom, 10)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1))),
				CreationHeight:   10,
			},
		},
	}

	for _, testCase := range testCases {
		suite.T().Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(suite.ctx, &testCase.req)
			if testCase.ExceptErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				balanceForNotBondedPool := suite.bankKeeper.GetBalance(suite.ctx, sdk.AccAddress(notBondedPool.GetAddress()), bondDenom)
				suite.Require().Equal(balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}
