package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
)

func (suite *KeeperTestSuite) TestHandleDoubleSign() {
	ctx := suite.ctx.WithIsCheckTx(false).WithBlockHeight(1)
	suite.populateValidators(ctx)

	power := int64(100)
	stakingParams := suite.stakingKeeper.GetParams(ctx)
	operatorAddr, val := valAddresses[0], pubkeys[0]
	tstaking := teststaking.NewHelper(suite.T(), ctx, suite.stakingKeeper)

	selfDelegation := tstaking.CreateValidatorWithValPower(operatorAddr, val, power, true)

	// execute end-blocker and verify validator attributes
	staking.EndBlocker(ctx, suite.stakingKeeper)
	suite.Equal(
		suite.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)).String(),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(selfDelegation))).String(),
	)
	suite.Equal(selfDelegation, suite.stakingKeeper.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	suite.slashingKeeper.HandleValidatorSignature(ctx, val.Address(), selfDelegation.Int64(), true)

	// double sign less than max age
	oldTokens := suite.stakingKeeper.Validator(ctx, operatorAddr).GetTokens()
	evidence := &types.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(val.Address()).String(),
	}
	suite.evidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	// should be jailed and tombstoned
	suite.True(suite.stakingKeeper.Validator(ctx, operatorAddr).IsJailed())
	suite.True(suite.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(val.Address())))

	// tokens should be decreased
	newTokens := suite.stakingKeeper.Validator(ctx, operatorAddr).GetTokens()
	suite.True(newTokens.LT(oldTokens))

	// submit duplicate evidence
	suite.evidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	// tokens should be the same (capped slash)
	suite.True(suite.stakingKeeper.Validator(ctx, operatorAddr).GetTokens().Equal(newTokens))

	// jump to past the unbonding period
	ctx = ctx.WithBlockTime(time.Unix(1, 0).Add(stakingParams.UnbondingTime))

	// require we cannot unjail
	suite.Error(suite.slashingKeeper.Unjail(ctx, operatorAddr))

	// require we be able to unbond now
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del, _ := suite.stakingKeeper.GetDelegation(ctx, sdk.AccAddress(operatorAddr), operatorAddr)
	validator, _ := suite.stakingKeeper.GetValidator(ctx, operatorAddr)
	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	tstaking.Ctx = ctx
	tstaking.Denom = stakingParams.BondDenom
	tstaking.Undelegate(sdk.AccAddress(operatorAddr), operatorAddr, totalBond, true)

	// query evidence from store
	evidences := suite.evidenceKeeper.GetAllEvidence(ctx)
	suite.Len(evidences, 1)
}

func (suite *KeeperTestSuite) TestHandleDoubleSign_TooOld() {
	ctx := suite.ctx.WithIsCheckTx(false).WithBlockHeight(1).WithBlockTime(time.Now())
	suite.populateValidators(ctx)

	power := int64(100)
	stakingParams := suite.stakingKeeper.GetParams(ctx)
	operatorAddr, val := valAddresses[0], pubkeys[0]
	tstaking := teststaking.NewHelper(suite.T(), ctx, suite.stakingKeeper)

	amt := tstaking.CreateValidatorWithValPower(operatorAddr, val, power, true)

	// execute end-blocker and verify validator attributes
	staking.EndBlocker(ctx, suite.stakingKeeper)
	suite.Equal(
		suite.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(amt))),
	)
	suite.Equal(amt, suite.stakingKeeper.Validator(ctx, operatorAddr).GetBondedTokens())

	evidence := &types.Equivocation{
		Height:           0,
		Time:             ctx.BlockTime(),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(val.Address()).String(),
	}

	cp := suite.app.BaseApp.GetConsensusParams(ctx)

	ctx = ctx.WithConsensusParams(cp)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(cp.Evidence.MaxAgeDuration + 1))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + cp.Evidence.MaxAgeNumBlocks + 1)
	suite.evidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	suite.False(suite.stakingKeeper.Validator(ctx, operatorAddr).IsJailed())
	suite.False(suite.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(val.Address())))
}
