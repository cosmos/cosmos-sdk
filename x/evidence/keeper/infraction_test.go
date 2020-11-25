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
	stakingParams := suite.app.StakingKeeper.GetParams(ctx)
	operatorAddr, val := valAddresses[0], pubkeys[0]
	tstaking := teststaking.NewHelper(suite.T(), ctx, suite.app.StakingKeeper)

	selfDelegation := tstaking.CreateValidatorWithValPower(operatorAddr, val, power, true)

	// execute end-blocker and verify validator attributes
	staking.EndBlocker(ctx, suite.app.StakingKeeper)
	suite.Equal(
		suite.app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)).String(),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(selfDelegation))).String(),
	)
	suite.Equal(selfDelegation, suite.app.StakingKeeper.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	suite.app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), selfDelegation.Int64(), true)

	// double sign less than max age
	oldTokens := suite.app.StakingKeeper.Validator(ctx, operatorAddr).GetTokens()
	evidence := &types.Equivocation{
		Height:           0,
		Time:             time.Unix(0, 0),
		Power:            power,
		ConsensusAddress: sdk.ConsAddress(val.Address()).String(),
	}
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	// should be jailed and tombstoned
	suite.True(suite.app.StakingKeeper.Validator(ctx, operatorAddr).IsJailed())
	suite.True(suite.app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(val.Address())))

	// tokens should be decreased
	newTokens := suite.app.StakingKeeper.Validator(ctx, operatorAddr).GetTokens()
	suite.True(newTokens.LT(oldTokens))

	// submit duplicate evidence
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	// tokens should be the same (capped slash)
	suite.True(suite.app.StakingKeeper.Validator(ctx, operatorAddr).GetTokens().Equal(newTokens))

	// jump to past the unbonding period
	ctx = ctx.WithBlockTime(time.Unix(1, 0).Add(stakingParams.UnbondingTime))

	// require we cannot unjail
	suite.Error(suite.app.SlashingKeeper.Unjail(ctx, operatorAddr))

	// require we be able to unbond now
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	del, _ := suite.app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(operatorAddr), operatorAddr)
	validator, _ := suite.app.StakingKeeper.GetValidator(ctx, operatorAddr)
	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	tstaking.Ctx = ctx
	tstaking.Denom = stakingParams.BondDenom
	tstaking.Undelegate(sdk.AccAddress(operatorAddr), operatorAddr, totalBond, true)
}

func (suite *KeeperTestSuite) TestHandleDoubleSign_TooOld() {
	ctx := suite.ctx.WithIsCheckTx(false).WithBlockHeight(1).WithBlockTime(time.Now())
	suite.populateValidators(ctx)

	power := int64(100)
	stakingParams := suite.app.StakingKeeper.GetParams(ctx)
	operatorAddr, val := valAddresses[0], pubkeys[0]
	tstaking := teststaking.NewHelper(suite.T(), ctx, suite.app.StakingKeeper)

	amt := tstaking.CreateValidatorWithValPower(operatorAddr, val, power, true)

	// execute end-blocker and verify validator attributes
	staking.EndBlocker(ctx, suite.app.StakingKeeper)
	suite.Equal(
		suite.app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, initAmt.Sub(amt))),
	)
	suite.Equal(amt, suite.app.StakingKeeper.Validator(ctx, operatorAddr).GetBondedTokens())

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
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(ctx, evidence)

	suite.False(suite.app.StakingKeeper.Validator(ctx, operatorAddr).IsJailed())
	suite.False(suite.app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(val.Address())))
}
